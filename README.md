# 服务计算-CLI开发|selpg

## 实验要求

使用 golang 开发 [开发 Linux 命令行实用程序](https://www.ibm.com/developerworks/cn/linux/shell/clutil/index.html) 中的 **selpg**

提示：

1. 请按文档 **使用 selpg** 章节要求测试你的程序
2. 请使用 pflag 替代 goflag 以满足 Unix 命令行规范， 参考：[Golang之使用Flag和Pflag](https://o-my-chenjian.com/2017/09/20/Using-Flag-And-Pflag-With-Golang/)
3. golang 文件读写、读环境变量，请自己查 os 包
4. “-dXXX” 实现，请自己查 `os/exec` 库，例如案例 [Command](https://godoc.org/os/exec#example-Command)，管理子进程的标准输入和输出通常使用 `io.Pipe`，具体案例见 [Pipe](https://godoc.org/io#Pipe)

在 Github 提交程序，并在 readme.md 文件中描述设计说明，使用与测试结果。

## 实验过程

### 设计说明

selpg 是从文本输入选择页范围的实用程序。该输入可以来自作为最后一个命令行参数指定的文件，在没有给出文件名参数时也可以来自标准输入。

selpg 首先处理所有的命令行参数。在扫描了所有的选项参数（也就是那些以连字符为前缀的参数）后，如果 selpg 发现还有一个参数，则它会接受该参数为输入文件的名称并尝试打开它以进行读取。如果没有其它参数，则 selpg 假定输入来自标准输入。

参考[selpg.c](<https://www.ibm.com/developerworks/cn/linux/shell/clutil/selpg.c>)

这里使用pflag包简化了参数的初始化.

使用```go get github.com/spf13/pflag```即可在线获取pflag包并在代码中```import flag "github.com/spf13/pflag"```这样就可以

新建一个结构体存储指令中需要用到的参数信息

```go
type selpg_Args struct {
	start, end, lines        int  
	readType, inputType, des string 
}
```

在FlagInit函数中让参数绑定变量

```go
	flag.IntVar(&args.start, "s", 0, "the start Page")             //开始页码，默认为0
	flag.IntVar(&args.end, "e", 0, "the end Page")                 //结束页码，默认为0
	flag.IntVar(&args.lines, "l", 72, "the lines of the page")     //每页行数，默认为72行每页
	flag.StringVar(&args.des, "d", "", "the destination to print") //输出位置，默认为空字符
```

这样输入命令测试的时候--s后的数字会赋值给args.start,没有设置就默认为0,其他参数同理

FlagInit后判断参数的合法性,如果不合法打印selpg用法提示信息并退出

```go
func usage() { //用法提示
	fmt.Fprintf(os.Stderr, "\nUSAGE: ./selpg [--s start] [--e end] [--l lines | --f ] [ --d des ] [ in_filename ]\n")
	fmt.Fprintf(os.Stderr, "\n selpg --s start    : start page")
	fmt.Fprintf(os.Stderr, "\n selpg --e end      : end page")
	fmt.Fprintf(os.Stderr, "\n selpg --l lines    : lines/page")
	fmt.Fprintf(os.Stderr, "\n selpg --f          : check page with '\\f'")
	fmt.Fprintf(os.Stderr, "\n selpg --d des      : pipe destination\n")
}
```

执行指令涉及到os包,os/exec包等调用

文件输入:

```go
if args.inputType != "" {
		fin, err = os.Open(args.inputType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Can't find input file \"%s\"!\n", args.inputType)
			usage()
			os.Exit(1)
		}
		defer fin.Close()
	}
```

管道输入:

```go
if args.des != "" {
		cmd := exec.Command("grep", "-nf", "keyword")
		inpipe, err = cmd.StdinPipe()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer inpipe.Close()
		cmd.Stdout = fout
		cmd.Start()
	}
```

然后使用bufio和io进行文件读写操作,扫描读入数据并用两个变量currentLine和currentPage作为计数器记录行数和页数

```go
	if args.readType == "l" {
		line := bufio.NewScanner(fin)
		for line.Scan() {
			if currentPage >= args.start && currentPage <= args.end {
				fout.Write([]byte(line.Text() + "\n"))
				if args.des != "" {
					inpipe.Write([]byte(line.Text() + "\n"))
				}
			}
			currentLine++
			if currentLine == args.lines {
				currentPage++
				currentLine = 0
			}
		}
	} else {
		read := bufio.NewReader(fin)
		for {
			page, ferr := read.ReadString('\f')
			if ferr != nil || ferr == io.EOF {
				if ferr == io.EOF {
					if currentPage >= args.start && currentPage <= args.end {
						fmt.Fprintf(fout, "%s", page)
					}
				}
				break
			}
			page = strings.Replace(page, "\f", "", -1)
			currentPage++
			if currentPage >= args.start && currentPage <= args.end {
				fmt.Fprintf(fout, "%s", page)
			}
		}
	}
	if currentPage < args.end {
		fmt.Fprintf(os.Stderr, "./selpg: end_page (%d) greater than total pages (%d), less output than expected\n", args.end, currentPage)
	}
}
```

在main函数中初始化一个args结构体并用FlagInit初始化,然后调用run函数执行指令

```c
func main() {
	var args selpg_Args
	FlagInit(&args)
	run(&args)
}
```

### 使用与测试

在selpg目录下打开终端使用```go build selpg.go```命令就可以在当前目录下生成可执行文件selpg然后用./selpg的方式执行命令.或者也可以用go install在bin目录下生成selpg可执行文件然后用selpg+参数的方式执行命令.接下来就可以进行测试了.

将文档中使用selpg的命令 的flag的``` - ```换成pflag的``` -- ```即可

测试文件采用c文件读写函数生成

先要搭建一下centos的c/c++编译环境

使用```sudo yum -y install gcc```安装c编译环境

使用```sudo yum -y install gcc-c++```安装c++编译环境

```c
#include <stdio.h>

int main()
{
	FILE *fp=fopen("input.txt","w");
	for(int i=1;i<=1000;i++){
		fprintf(fp,"%d\n",i);
		if(i%10==0){
			fprintf(fp,"\f");
		}
	}
	fclose(fp);
	return 0;
 } 
```

生成一个每行有行号,每10行一个换页符"\f"的txt文件,总共设定了1000行,所以有100页.但是以72行为一页的话不足20页

终端输入命令

![img/1569483060374](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569483060374.png)

或者

![img/1569483175214](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569483175214.png)

生成的input.txt如下:

![img/1569458832310](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569458832310.png)

`selpg` 命令字符串示例：

1. `$ ./selpg --s 1 --e 1 input.txt`

   该命令将把“input_file”的第 1 页写至标准输出（也就是屏幕），因为这里没有重定向或管道。

   ![img/1569483409603](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569483409603.png)

   ![img/1569483442989](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569483442989.png)

   1页72行,结果正确

2. `$ ./selpg --s 1 --e 1 < input.txt`

   该命令与示例 1 所做的工作相同，但在本例中，selpg 读取标准输入，而标准输入已被 shell／内核重定向为来自“input_file”而不是显式命名的文件名参数。输入的第 1 页被写至屏幕。

   ![img/1569483543052](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569483543052.png)

   结果同1

3. `$ ./selpg --s 10 --e 20 input.txt >output.txt`

   selpg 将第 10 页到第 20 页写至标准输出；标准输出被 shell／内核重定向至“output_file”。

   ![img/1569483829161](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569483829161.png)

   至input.txt尾1000行

4. `$ ./selpg --s 10 --e 20 input.txt 2>error.txt`

   selpg 将第 10 页到第 20 页写至标准输出（屏幕）；所有的错误消息被 shell／内核重定向至“error_file”。请注意：在“2”和“>”之间不能有空格；这是 shell 语法的一部分（请参阅“man bash”或“man sh”）。

   ![img/1569484077675](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569484077675.png)

   ![img/1569484087645](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569484087645.png)

   错误信息为空

   将s 10换成s -10

   ![img/1569494740117](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569494740117.png)

   错误信息输出到error.txt

5. `$ ./selpg --s 10 --e 20 input.txt >output2.txt 2>error.txt`

   selpg 将第 10 页到第 20 页写至标准输出，标准输出被重定向至“output_file”；selpg 写至标准错误的所有内容都被重定向至“error_file”。当“input_file”很大时可使用这种调用；您不会想坐在那里等着 selpg 完成工作，并且您希望对输出和错误都进行保存。

   ![img/1569495682908](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569495682908.png)

   error.txt被置空,结果正确

6. `$ ./selpg --s 3 --e 5 input.txt >output2.txt 2>/dev/null`

   selpg 将第 10 页到第 20 页写至标准输出，标准输出被重定向至“output_file”；selpg 写至标准错误的所有内容都被重定向至 /dev/null（空设备），这意味着错误消息被丢弃了。设备文件 /dev/null 废弃所有写至它的输出，当从该设备文件读取时，会立即返回 EOF。

   结果同上.

   把参数s 3换成-1

   ![img/1569498221331](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569498221331.png)

   ![img/1569496510105](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569496510105.png)

   output2.txt文件也被清空了

7. `$ ./selpg --s 10 --e 20 input.txt >/dev/null`

   selpg 将第 10 页到第 20 页写至标准输出，标准输出被丢弃；错误消息在屏幕出现。这可作为测试 selpg 的用途，此时您也许只想（对一些测试情况）检查错误消息，而不想看到正常输出。

![img/1569484474395](../../../../../../%E5%AD%A6%E4%B9%A0/%E8%AF%BE%E4%B8%9A/%E6%9C%8D%E5%8A%A1%E8%AE%A1%E7%AE%97/homework/hw4/selpg/img/1569484474395.png)

将20换成-20,错误信息会在屏幕上出现,结果正确
