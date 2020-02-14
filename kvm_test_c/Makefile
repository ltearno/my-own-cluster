all: kvm_test

kvm_test: kvm_test.c
	gcc kvm_test.c -o kvm_test

kvm_bare: kvm_bare.c
	gcc kvm_bare.c -o kvm_bare

bare: bare.c
	cc -c bare.c
	objcopy bare.o bare -O binary