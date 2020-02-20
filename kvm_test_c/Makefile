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

.PHONY: disasm
disasm: helloworld-gas-16bit.bin
	objdump -D helloworld-gas-16bit.o
	hexdump -C helloworld-gas-16bit.bin

.PHONY: run-kvm_bare
run-kvm_bare: kvm_bare disasm helloworld-gas-16bit.bin
	./kvm_bare helloworld-gas-16bit.bin