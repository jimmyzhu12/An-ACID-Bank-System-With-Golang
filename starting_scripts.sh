go run ./cmd/branch A ./config/sample_config.txt
go run ./cmd/branch B ./config/sample_config.txt
go run ./cmd/branch C ./config/sample_config.txt
go run ./cmd/branch D ./config/sample_config.txt
go run ./cmd/branch E ./config/sample_config.txt

go run ./cmd/client abcd ./config/sample_config.txt
go run ./cmd/client efg ./config/sample_config.txt
go run ./cmd/client hik ./config/sample_config.txt

BEGIN
DEPOSIT B.foo 10
BALANCE B.foo

./server A ./config/vm_config.txt
./server B ./config/vm_config.txt
./server C ./config/vm_config.txt
./server D ./config/vm_config.txt
./server E ./config/vm_config.txt

./client abcd ./config/vm_config.txt
./client efg ./config/vm_config.txt
./client hjk ./config/vm_config.txt