use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SmtpNotifierConfig {
    pub host: String,
    pub port: i32,
    pub username: String,
    pub password: String,
    pub sender: String,
    pub recipients: Vec<String>,
    pub subject_template: String,
    pub body_template: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiscordWebhookNotifierReuseMessageConfig {
    pub enable: bool,
    pub message_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiscordWebhookNotifierConfig {
    pub url: String,
    pub message_template: String,
    pub reuse_message: Option<DiscordWebhookNotifierReuseMessageConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GotifyNotifierConfig {
    pub url: String,
    pub token: String,
    pub title_template: String,
    pub message_template: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NotifierConfig {
    pub r#type: String,
    pub config: serde_json::Value,
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

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RetryConfiguration {
    pub max_attempts: i32,
    pub interval: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MonitorConfig {
    pub name: String,
    pub interval: String,
    pub probe: ProbeConfig,
    pub notifiers: Vec<String>,
    pub consider_all_tests: bool,
    pub invert: bool,
    pub retry: RetryConfiguration,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MetricsConfiguration {
    pub enable: bool,
    pub listen_address: String,
    pub extra_labels: std::collections::HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemberlistConfiguration {
    pub bind_address: String,
    pub bind_port: i32,
    pub advertise_address: String,
    pub advertise_port: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClusterConfiguration {
    pub enable: bool,
    pub memberlist: MemberlistConfiguration,
    pub join: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Configuration {
    pub instance: String,
    pub metrics: MetricsConfiguration,
    pub cluster: ClusterConfiguration,
    pub concurrent_tasks: i32,
    pub notifiers: std::collections::HashMap<String, NotifierConfig>,
    pub include_notifiers: Option<String>,
    pub monitors: Vec<MonitorConfig>,
    pub include_monitors: Option<String>,
}
