PAYLOAD64 := helloworld-gas-64bit

all: kvm_test

kvm_test: kvm_test.c
	gcc kvm_test.c -o kvm_test

kvm_bare: kvm_bare.c
	gcc kvm_bare.c -o kvm_bare

bare: bare.c
	cc -c bare.c
	objcopy bare.o bare -O binary

helloworld-gas-16bit.o: helloworld-gas-16bit.s
	as -32 helloworld-gas-16bit.s -o helloworld-gas-16bit.o

helloworld-gas-16bit.bin: helloworld-gas-16bit.o
	ld -m elf_i386 --oformat binary -N -e _start -Ttext 0x10000 -o helloworld-gas-16bit.bin helloworld-gas-16bit.o

.PHONY: disasm-16
disasm-16: helloworld-gas-16bit.bin
	objdump -D helloworld-gas-16bit.o
	hexdump -C helloworld-gas-16bit.bin

$(PAYLOAD64).o: $(PAYLOAD64).s
	as -64 $(PAYLOAD64).s -o $(PAYLOAD64).o

$(PAYLOAD64).bin: $(PAYLOAD64).o
	ld --oformat binary -N -e _start -Ttext 0x10000 -o $(PAYLOAD64).bin $(PAYLOAD64).o

.PHONY: disasm-64
disasm-64: $(PAYLOAD64).bin
	objdump -D $(PAYLOAD64).o
	hexdump -C $(PAYLOAD64).bin

.PHONY: run-kvm_bare
run-kvm_bare-16: kvm_bare disasm-16 helloworld-gas-16bit.bin
	./kvm_bare helloworld-gas-16bit.bin

run-kvm_bare-64: kvm_bare disasm-64 $(PAYLOAD64).bin
	./kvm_bare $(PAYLOAD64).bin