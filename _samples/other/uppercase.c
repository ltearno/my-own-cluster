#include <stdio.h>
#include <ctype.h>

#define BUFSIZE 1024

int main(int argc, char* argv[]) {
    char buf[BUFSIZE+1];

    while(1){
        // read stdin
		int readden = fread(buf, 1, BUFSIZE, stdin);
		buf[readden] = 0;

        // upper case
        for(int i=0; i<readden; i++)
            buf[i] = toupper(buf[i]);
        
        // write stdout
        fwrite(buf, 1, readden, stdout);
        
		if(readden != BUFSIZE)
			break;
	}

    return 0;
}