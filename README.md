# Scalable Worker Pool

## Project Overview
Scalable Worker Pool is a dynamically scalable worker pool management system. This application utilizes workers to process incoming requests, providing high efficiency and flexibility. The application automatically increases and decreases the number of workers to optimize system load.

### Key Features
- Dynamic worker management
- Load monitoring and automatic scaling
- Request processing logic
- Graceful shutdown support

## Project Structure
The project is structured according to Clean Architecture principles. Below is the main components of the project along with the flow diagram.

### Flow Diagram
```plaintext
+---------------------+
|      main.go       |
+---------------------+
          |
          | Initializes context, sets max procs, and creates request handler
          |
          v
+---------------------+
|   NewDispatcher     |
|  (workerpool)       |
+---------------------+
          |
          | Creates WorkerManager and Scaler
          |
          v
+---------------------+
|  WorkerManager      |
+---------------------+
          |
          | Manages workers and their lifecycle
          |
          v
+---------------------+
|      Worker         |
+---------------------+
          |
          | Launches worker to process requests
          |
          v
+---------------------+
|      Request        |
+---------------------+
          |
          | Contains request data and processing logic
          |
          v
+---------------------+
|   RequestHandler    |
+---------------------+
          |
          | Handles specific request types
          |
          v
+---------------------+
|      Scaler         |
+---------------------+
          |
          | Monitors load and scales workers up/down
          |
          v
+---------------------+
|   Dispatcher        |
+---------------------+
          |
          | Manages the overall worker pool and request distribution
          |
          v
+---------------------+
|   Graceful Shutdown  |
+---------------------+
          |
          | Ensures all requests are processed before exiting
          |
          v
+---------------------+
|      Exit           |
+---------------------+
```

## Installation
1. Ensure that Go is installed on your machine. You can download it from the official Go website: [golang.org](https://golang.org/dl/).
2. Clone the repository to your local machine:
   ```bash
   git clone https://github.com/cembdc/scalable-worker-pool
   ```
3. Navigate to the project directory:
   ```bash
   cd scalable-worker-pool
   ```
4. Install the necessary dependencies using Go modules:
   ```bash
   go mod tidy
   ```

## Usage
To start the application, run the following command:
```bash
go run main.go
```

## Contributing
We welcome contributions to this project! If you have suggestions or improvements, please feel free to submit a pull request or open an issue.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
