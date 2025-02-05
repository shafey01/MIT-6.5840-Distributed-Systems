module github.com/shafey01/MIT-6.006-Algorithems/rpc/main

replace github.com/shafey01/MIT-6.006-Algorithems/rpc/client => /home/shafey/learning-go/MIT-6.5840-Distributed-Systems/rpc/client

replace github.com/shafey01/MIT-6.006-Algorithems/rpc/server => /home/shafey/learning-go/MIT-6.5840-Distributed-Systems/rpc/server

go 1.23.5

require github.com/shafey01/MIT-6.006-Algorithems/rpc/server v0.0.0-00010101000000-000000000000

require github.com/google/uuid v1.6.0 // indirect
