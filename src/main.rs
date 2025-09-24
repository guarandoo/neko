pub mod config;
pub mod core;
pub mod metrics;
pub mod monitor;
pub mod notifier;
pub mod probe;

use crate::config::Configuration;
use crate::metrics::SimpleMetricsServer;
use crate::monitor::Monitor;
use crate::notifier::{DiscordWebhookNotifier, GotifyNotifier, SmtpNotifier};
use crate::probe::{ExecProbe, HttpProbe, PingProbe};
use serde_yaml;
use std::fs;
use std::sync::{Arc, Mutex};

pub struct Application {
    pub metrics_server: SimpleMetricsServer,
    pub configuration_file: String,
    pub configuration: Option<Configuration>,
}

impl Application {
    pub fn new() -> Self {
        Application {
            metrics_server: SimpleMetricsServer::new(),
            configuration_file: String::new(),
            configuration: None,
        }
    }

    pub fn reload(&mut self) -> Result<(), String> {
        match Self::load_configuration(&self.configuration_file) {
            Ok(config) => {
                self.configuration = Some(config);
                Ok(())
            }
            Err(e) => Err(e),
        }
    }

    pub fn run(&mut self) -> Result<(), String> {
        println!("Running neko");
        self.configuration_file = "config.yaml".to_string();
        self.reload()?;
        println!("Loaded configuration: {:?}", self.configuration_file);

        let monitors = vec![
            Monitor {
                name: "exec_monitor".to_string(),
                interval: "5s".to_string(),
                probe: Box::new(ExecProbe::new("/bin/true".to_string(), vec![])),
                status: crate::core::Status::Pending,
                notifiers: vec![Box::new(SmtpNotifier::new())],
                configuration: self.configuration.as_ref().unwrap().monitors[0].clone(),
            },
            Monitor {
                name: "ping_monitor".to_string(),
                interval: "5s".to_string(),
                probe: Box::new(PingProbe::new("127.0.0.1".to_string(), 1)),
                status: crate::core::Status::Pending,
                notifiers: vec![Box::new(DiscordWebhookNotifier::new())],
                configuration: self.configuration.as_ref().unwrap().monitors[0].clone(),
            },
            Monitor {
                name: "http_monitor".to_string(),
                interval: "5s".to_string(),
                probe: Box::new(HttpProbe::new(
                    "http://localhost".to_string(),
                    "GET".to_string(),
                    5,
                    vec![200],
                    std::collections::HashMap::new(),
                )),
                status: crate::core::Status::Pending,
                notifiers: vec![Box::new(GotifyNotifier::new())],
                configuration: self.configuration.as_ref().unwrap().monitors[0].clone(),
            },
        ];

        let ctx = Arc::new(Mutex::new(()));
        for mut monitor in monitors {
            let ctx = ctx.clone();
            std::thread::spawn(move || {
                loop {
                    let result = monitor.probe.probe(&ctx, "instance", &monitor.name);
                    match result {
                        Ok(res) => {
                            let status = if res
                                .tests
                                .iter()
                                .any(|t| t.status == crate::core::Status::Up)
                            {
                                crate::core::Status::Up
                            } else {
                                crate::core::Status::Down
                            };
                            let status_val = status.clone();
                            monitor.status = status;
                            for notifier in &monitor.notifiers {
                                let mut data = std::collections::HashMap::new();
                                data.insert(
                                    "Name".to_string(),
                                    serde_json::Value::String(monitor.name.clone()),
                                );
                                data.insert(
                                    "Status".to_string(),
                                    serde_json::Value::String(format!("{:?}", status_val)),
                                );
                                let _ = notifier.notify(&ctx, &monitor.name, &data);
                            }
                        }
                        Err(e) => {
                            println!("Monitor {} probe error: {}", monitor.name, e);
                        }
                    }
                    std::thread::sleep(std::time::Duration::from_secs(5));
                }
            });
        }
        Ok(())
    }

    fn load_configuration(path: &str) -> Result<Configuration, String> {
        let text =
            fs::read_to_string(path).map_err(|e| format!("Unable to read config file: {}", e))?;
        serde_yaml::from_str(&text).map_err(|e| format!("Unable to parse config: {}", e))
    }
}

fn main() {
    println!("Starting neko");
    let mut app = Application::new();
    if let Err(e) = app.run() {
        eprintln!("Application error: {}", e);
    }
}
