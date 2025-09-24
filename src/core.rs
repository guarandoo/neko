#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Status {
    Pending,
    Up,
    Down,
}

impl std::fmt::Display for Status {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Status::Pending => write!(f, "Unknown"),
            Status::Up => write!(f, "Up"),
            Status::Down => write!(f, "Down"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct Test {
    pub target: String,
    pub status: Status,
    pub error: Option<String>,
    pub extras: std::collections::HashMap<String, serde_json::Value>,
}

#[derive(Debug, Clone)]
pub struct Result {
    pub tests: Vec<Test>,
}

#[derive(Debug, Clone)]
pub struct Datapoint {
    pub measured_at: std::time::SystemTime,
    pub status: Status,
}
