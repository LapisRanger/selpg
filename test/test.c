#include <stdio.h>

int main()
{
    FILE *fp=fopen("input.txt","w");
    for(int i=1;i<=1000;i++){
        fprintf(fp,"%d\n",i);
        if(i%10==0) fprintf(fp,"\f");
    }
    fclose(fp);
    return 0;
}