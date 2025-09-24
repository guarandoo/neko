pub trait MetricsServer {
    fn listen(&mut self, addr: &str) -> Result<(), String>;
    fn close(&mut self) -> Result<(), String>;
}

use std::io::{Read, Write};
use std::net::TcpListener;
use std::thread;

pub struct SimpleMetricsServer {
    running: bool,
}

impl SimpleMetricsServer {
    pub fn new() -> Self {
        SimpleMetricsServer { running: false }
    }
}

impl MetricsServer for SimpleMetricsServer {
    fn listen(&mut self, addr: &str) -> Result<(), String> {
        if self.running {
            return Ok(());
        }
        self.running = true;
        let addr = addr.to_string();
        thread::spawn(move || {
            let listener = TcpListener::bind(&addr).expect("Failed to bind metrics server");
            for stream in listener.incoming() {
                if let Ok(mut stream) = stream {
                    let mut buffer = [0; 512];
                    let _ = stream.read(&mut buffer);
                    let response =
                        "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nneko_metrics_stub 1\n";
                    let _ = stream.write(response.as_bytes());
                }
            }
        });
        Ok(())
    }
    fn close(&mut self) -> Result<(), String> {
        self.running = false;
        Ok(())
    }
}
