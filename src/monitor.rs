use crate::config::MonitorConfig;
use crate::core::Status;
use crate::notifier::Notifier;
use crate::probe::Probe;

pub struct Monitor {
    pub name: String,
    pub interval: String,
    pub probe: Box<dyn Probe + Send + Sync>,
    pub status: Status,
    pub notifiers: Vec<Box<dyn Notifier + Send + Sync>>,
    pub configuration: MonitorConfig,
}
