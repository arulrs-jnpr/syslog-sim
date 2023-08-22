# syslog-sim

Sends junos binary format syslogs to mist log-term over tls connection.

Build
go build ##generates binary executable file (./sim)

Usage 
go run syslog_generator.go [options]

or

./sim [options]

Supported Options:
-m #mac of internal config to be retrived
-c #device internal config json file>
-p #number of parallel connections to create
-i #interval between the iterations in seconds

Example:
./sim -c ./conf/4c9614c95000.json -p 10 -i 30
