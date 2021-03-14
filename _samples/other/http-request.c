#include <errno.h>
#include <stdio.h>
#include <dirent.h>
#include <stdlib.h>
#include <sys/stat.h>

#define BUFSIZE 1024

int main(int argc, char* argv[]) {
	printf("stdin is %d\n", stdin);
	printf("stdout is %d\n", stdout);
	printf("stderr is %d\n", stderr);

	if(argc<2) {
		printf("YOU SHOULD GIVE A PARAMETER (THE URL)\n");
		return 1;
	}

	char buf[BUFSIZE+1];
	buf[0] = 0;

	errno = 0;
	int fd = fopen(argv[1], "rw");
	printf("opened %d (err %d)\n", fd, errno);
	fflush(stdout);

	int fdOutput = fopen("api://output", "w");
	printf("opened OUTPUT %d (err %d)\n", fdOutput, errno);


	int total = 0;
	printf("RECEIVING ... :\n");
	while(1){
		int readden = fread(buf, 1, BUFSIZE, fd);
		fwrite( buf,1, readden, fdOutput);
		fflush(fdOutput);
		total += readden;
		buf[readden] = 0;
		printf("%s", buf);
		if(readden != BUFSIZE)
			break;
	}
	printf(">>>> that was %d bytes !\n", total);

	fclose(fdOutput);

	fd = fopen("api://input", "r");
	while(1){
		int readden = fread(buf, 1, BUFSIZE, fd);
		buf[readden] = 0;
		printf("%s", buf);
		if(readden != BUFSIZE)
			break;
	}

	printf("\nREADING STDIN\n");
	while(1){
		int readden = fread(buf, 1, BUFSIZE, stdin);
		buf[readden] = 0;
		printf("from_stdin): %d %s\n", readden, buf);
		if(readden != BUFSIZE)
			break;
	}

	printf("KJHGDKJ HDGKJDH GKJss HDGKJDG HKDJGH DKJGHD\n");

	return 9;
}
