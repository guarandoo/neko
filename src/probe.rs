use crate::core::Result as ProbeResult;
use crate::core::{Status, Test};
use chrono::{DateTime, Utc};
use ping::ping;
use regex::Regex;
use reqwest::blocking::Client;
use reqwest::header::{HeaderMap, HeaderName, HeaderValue};
use serde::{Deserialize, Serialize};
use ssh2::Session;
use std::collections::HashMap;
use std::net::TcpStream;
use std::process::Command;
use std::sync::{Arc, Mutex};
use trust_dns_resolver::config::*;
use trust_dns_resolver::Resolver;
use whois_rust::{WhoIs, WhoIsLookupOptions};

pub trait Probe {
    fn probe(
        &self,
        ctx: &Arc<Mutex<()>>,
        instance: &str,
        name: &str,
    ) -> Result<ProbeResult, String>;
}

pub struct ProbeOptions {}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecProbeTypeConfig {
    pub path: String,
    pub args: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PingProbeTypeConfig {
    pub address: String,
    pub count: i32,
    pub packet_loss_threshold: f64,
    pub privileged: bool,
    pub interval: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HttpProbeTypeConfig {
    pub address: String,
    pub socket_path: Option<String>,
    pub max_redirects: i32,
    pub success_status_codes: Vec<i32>,
    pub headers: std::collections::HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProbeConfig {
    pub r#type: String,
    pub timeout: u64,
    pub config: serde_json::Value,
}

pub struct ExecProbe {
    pub path: String,
    pub args: Vec<String>,
}

impl ExecProbe {
    pub fn new(path: String, args: Vec<String>) -> Self {
        ExecProbe { path, args }
    }
}

impl Probe for ExecProbe {
    fn probe(
        &self,
        _ctx: &std::sync::Arc<std::sync::Mutex<()>>,
        _instance: &str,
        name: &str,
    ) -> Result<ProbeResult, String> {
        let output = Command::new(&self.path).args(&self.args).output();
        let (status, error) = match &output {
            Ok(o) if o.status.success() => (Status::Up, None),
            Ok(_) => (Status::Down, None),
            Err(e) => (Status::Down, Some(format!("Exec error: {}", e))),
        };
        let test = Test {
            target: name.to_string(),
            status,
            error,
            extras: HashMap::new(),
        };
        Ok(ProbeResult { tests: vec![test] })
    }
}

pub struct PingProbe {
    pub address: String,
    pub count: u32,
}

impl PingProbe {
    pub fn new(address: String, count: u32) -> Self {
        PingProbe { address, count }
    }
}

impl Probe for PingProbe {
    fn probe(
        &self,
        _ctx: &std::sync::Arc<std::sync::Mutex<()>>,
        _instance: &str,
        name: &str,
    ) -> Result<ProbeResult, String> {
        let mut status = Status::Down;
        let mut error = None;
        let ip_addr = match self.address.parse() {
            Ok(addr) => addr,
            Err(e) => {
                error = Some(format!("Invalid IP address: {}", e));
                let test = Test {
                    target: name.to_string(),
                    status,
                    error,
                    extras: HashMap::new(),
                };
                return Ok(ProbeResult { tests: vec![test] });
            }
        };
        match ping(ip_addr, None, Some(self.count), None, Some(1), None) {
            Ok(_) => {
                status = Status::Up;
            }
            Err(e) => {
                error = Some(format!("Ping error: {}", e));
            }
        }
        let test = Test {
            target: name.to_string(),
            status,
            error,
            extras: HashMap::new(),
        };
        Ok(ProbeResult { tests: vec![test] })
    }
}

pub struct HttpProbe {
    pub url: String,
    pub method: String,
    pub max_redirects: i32,
    pub success_status_codes: Vec<i32>,
    pub headers: HashMap<String, String>,
}

impl HttpProbe {
    pub fn new(
        url: String,
        method: String,
        max_redirects: i32,
        success_status_codes: Vec<i32>,
        headers: HashMap<String, String>,
    ) -> Self {
        HttpProbe {
            url,
            method,
            max_redirects,
            success_status_codes,
            headers,
        }
    }
}

impl Probe for HttpProbe {
    fn probe(
        &self,
        _ctx: &std::sync::Arc<std::sync::Mutex<()>>,
        _instance: &str,
        name: &str,
    ) -> Result<ProbeResult, String> {
        let client = Client::builder()
            .redirect(reqwest::redirect::Policy::limited(
                self.max_redirects as usize,
            ))
            .build()
            .map_err(|e| format!("Failed to build HTTP client: {}", e))?;

        let mut req = client.request(
            self.method.parse().unwrap_or(reqwest::Method::GET),
            &self.url,
        );
        let mut header_map = HeaderMap::new();
        for (k, v) in &self.headers {
            if let (Ok(name), Ok(value)) = (
                HeaderName::from_bytes(k.as_bytes()),
                HeaderValue::from_str(v),
            ) {
                header_map.insert(name, value);
            }
        }
        req = req.headers(header_map);

        let mut test = Test {
            target: name.to_string(),
            status: Status::Up,
            error: None,
            extras: HashMap::new(),
        };

        match req.send() {
            Ok(res) => {
                let code = res.status().as_u16() as i32;
                if !self.success_status_codes.contains(&code) {
                    test.status = Status::Down;
                    test.error = Some(format!("return code was {}", code));
                }
            }
            Err(e) => {
                test.status = Status::Down;
                test.error = Some(format!("HTTP error: {}", e));
            }
        }
        Ok(ProbeResult { tests: vec![test] })
    }
}

pub struct DnsProbe {
    pub server: String,
    pub port: u16,
    pub target: String,
    pub record_type: String,
}

impl DnsProbe {
    pub fn new(server: String, port: u16, target: String, record_type: String) -> Self {
        DnsProbe {
            server,
            port,
            target,
            record_type,
        }
    }
}

impl Probe for DnsProbe {
    fn probe(
        &self,
        _ctx: &std::sync::Arc<std::sync::Mutex<()>>,
        _instance: &str,
        name: &str,
    ) -> Result<ProbeResult, String> {
        let mut test = Test {
            target: name.to_string(),
            status: Status::Up,
            error: None,
            extras: HashMap::new(),
        };
        let resolver = Resolver::new(ResolverConfig::default(), ResolverOpts::default())
            .map_err(|e| format!("DNS resolver error: {}", e))?;
        let result = match self.record_type.as_str() {
            "A" | "Host" => resolver.lookup_ip(&self.target).map(|_| ()),
            "NS" => resolver
                .lookup(&self.target, trust_dns_resolver::proto::rr::RecordType::NS)
                .map(|_| ()),
            "MX" => resolver
                .lookup(&self.target, trust_dns_resolver::proto::rr::RecordType::MX)
                .map(|_| ()),
            "CNAME" => resolver
                .lookup(
                    &self.target,
                    trust_dns_resolver::proto::rr::RecordType::CNAME,
                )
                .map(|_| ()),
            _ => Err(trust_dns_resolver::error::ResolveErrorKind::Msg(
                "Unknown record type".to_string(),
            )
            .into()),
        };
        match result {
            Ok(_) => {}
            Err(e) => {
                test.status = Status::Down;
                test.error = Some(format!("DNS error: {}", e));
            }
        }
        Ok(ProbeResult { tests: vec![test] })
    }
}

pub struct SshProbe {
    pub host: String,
    pub port: u16,
    pub username: String,
    pub password: Option<String>,
    pub privkey: Option<String>,
}

impl SshProbe {
    pub fn new(
        host: String,
        port: u16,
        username: String,
        password: Option<String>,
        privkey: Option<String>,
    ) -> Self {
        SshProbe {
            host,
            port,
            username,
            password,
            privkey,
        }
    }
}

impl Probe for SshProbe {
    fn probe(
        &self,
        _ctx: &std::sync::Arc<std::sync::Mutex<()>>,
        _instance: &str,
        name: &str,
    ) -> Result<ProbeResult, String> {
        let mut test = Test {
            target: name.to_string(),
            status: Status::Down,
            error: None,
            extras: HashMap::new(),
        };
        let tcp = TcpStream::connect(format!("{}:{}", self.host, self.port));
        match tcp {
            Ok(stream) => {
                let mut session =
                    Session::new().map_err(|e| format!("SSH session error: {}", e))?;
                session.set_tcp_stream(stream);
                session
                    .handshake()
                    .map_err(|e| format!("SSH handshake error: {}", e))?;
                let auth_result = if let Some(ref pw) = self.password {
                    session.userauth_password(&self.username, pw)
                } else if let Some(ref pk) = self.privkey {
                    session.userauth_pubkey_file(
                        &self.username,
                        None,
                        std::path::Path::new(pk),
                        None,
                    )
                } else {
                    Err(ssh2::Error::from_errno(ssh2::ErrorCode::Session(-1)))
                };
                match auth_result {
                    Ok(_) => {
                        if session.authenticated() {
                            test.status = Status::Up;
                        } else {
                            test.error = Some("SSH authentication failed".to_string());
                        }
                    }
                    Err(e) => {
                        test.error = Some(format!("SSH auth error: {}", e));
                    }
                }
            }
            Err(e) => {
                test.error = Some(format!("SSH TCP error: {}", e));
            }
        }
        Ok(ProbeResult { tests: vec![test] })
    }
}

pub struct DomainProbe {
    pub domain: String,
    pub threshold_hours: u64,
}

impl DomainProbe {
    pub fn new(domain: String, threshold_hours: u64) -> Self {
        DomainProbe {
            domain,
            threshold_hours,
        }
    }
}

impl Probe for DomainProbe {
    fn probe(
        &self,
        _ctx: &std::sync::Arc<std::sync::Mutex<()>>,
        _instance: &str,
        name: &str,
    ) -> Result<ProbeResult, String> {
        let mut test = Test {
            target: name.to_string(),
            status: Status::Down,
            error: None,
            extras: HashMap::new(),
        };
        let whois = WhoIs::from_string("whois.verisign-grs.com")
            .unwrap_or_else(|_| WhoIs::from_string("").unwrap());
        let opts = match WhoIsLookupOptions::from_string(&self.domain) {
            Ok(o) => o,
            Err(e) => {
                test.error = Some(format!("WhoIs options error: {}", e));
                return Ok(ProbeResult { tests: vec![test] });
            }
        };
        match whois.lookup(opts) {
            Ok(response) => {
                let re = Regex::new(r"(?i)Expiry Date:\s*(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)")
                    .unwrap();
                if let Some(caps) = re.captures(&response) {
                    if let Some(date_str) = caps.get(1) {
                        if let Ok(expiry) = DateTime::parse_from_rfc3339(date_str.as_str()) {
                            let now = Utc::now();
                            let remaining = expiry.with_timezone(&Utc) - now;
                            let hours = remaining.num_hours();
                            test.extras.insert(
                                "remaining_hours".to_string(),
                                serde_json::Value::Number(hours.into()),
                            );
                            if hours > self.threshold_hours as i64 {
                                test.status = Status::Up;
                            } else {
                                test.error = Some(format!("Domain expires in {} hours", hours));
                            }
                        } else {
                            test.error = Some("Could not parse expiry date".to_string());
                        }
                    }
                } else {
                    test.error = Some("Expiry date not found in WHOIS response".to_string());
                }
            }
            Err(e) => {
                test.error = Some(format!("WHOIS error: {}", e));
            }
        }
        Ok(ProbeResult { tests: vec![test] })
    }
}
