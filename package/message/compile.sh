# Run this line if it says:
# protoc-gen-go: program not found or is not executable
export PATH="$PATH:$(go env GOPATH)/bin"

protoc --go_out=. --go_opt=paths=source_relative ./package/message/command.proto
protoc --go_out=. --go_opt=paths=source_relative ./package/message/command_res.proto
protoc --go_out=. --proto_path=. --go_opt=paths=source_relative ./package/message/branch_msg.proto