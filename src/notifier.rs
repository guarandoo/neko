pub trait Notifier {
    fn notify(
        &self,
        ctx: &std::sync::Arc<std::sync::Mutex<()>>,
        name: &str,
        data: &std::collections::HashMap<String, serde_json::Value>,
    ) -> Result<(), String>;
}

#[derive(Debug, Clone)]
pub struct SmtpNotifierOptions {
    pub host: String,
    pub port: i32,
    pub username: String,
    pub password: String,
    pub sender: String,
    pub recipients: Vec<String>,
    pub subject_template: String,
    pub body_template: String,
}

#[derive(Debug, Clone)]
pub struct DiscordWebhookOptions {
    pub url: String,
    pub message_template: String,
    pub persistent_message: bool,
    pub last_message_id: Option<String>,
}

#[derive(Debug, Clone)]
pub struct GotifyOptions {
    pub url: String,
    pub token: String,
    pub title_template: String,
    pub message_template: String,
}

pub struct SmtpNotifier;
impl SmtpNotifier {
    pub fn new() -> Self {
        SmtpNotifier
    }
}

impl Notifier for SmtpNotifier {
    fn notify(
        &self,
        _ctx: &std::sync::Arc<std::sync::Mutex<()>>,
        name: &str,
        data: &std::collections::HashMap<String, serde_json::Value>,
    ) -> Result<(), String> {
        println!("SMTP Notifier: {} -> {:?}", name, data);
        Ok(())
    }
}

pub struct DiscordWebhookNotifier;
impl DiscordWebhookNotifier {
    pub fn new() -> Self {
        DiscordWebhookNotifier
    }
}

impl Notifier for DiscordWebhookNotifier {
    fn notify(
        &self,
        _ctx: &std::sync::Arc<std::sync::Mutex<()>>,
        name: &str,
        data: &std::collections::HashMap<String, serde_json::Value>,
    ) -> Result<(), String> {
        println!("Discord Webhook Notifier: {} -> {:?}", name, data);
        Ok(())
    }
}

pub struct GotifyNotifier;
impl GotifyNotifier {
    pub fn new() -> Self {
        GotifyNotifier
    }
}

impl Notifier for GotifyNotifier {
    fn notify(
        &self,
        _ctx: &std::sync::Arc<std::sync::Mutex<()>>,
        name: &str,
        data: &std::collections::HashMap<String, serde_json::Value>,
    ) -> Result<(), String> {
        println!("Gotify Notifier: {} -> {:?}", name, data);
        Ok(())
    }
}
