package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	flag "github.com/spf13/pflag"
)

type selpg_Args struct {
	start, end, lines        int    //起始页码,终止页码,行数
	readType, inputType, des string //读取方式，输入方式，输出位置
}

func usage() { //用法提示
	fmt.Fprintf(os.Stderr, "\nUSAGE: ./selpg [--s start] [--e end] [--l lines | --f ] [ --d des ] [ in_filename ]\n")
	fmt.Fprintf(os.Stderr, "\n selpg --s start    : start page")
	fmt.Fprintf(os.Stderr, "\n selpg --e end      : end page")
	fmt.Fprintf(os.Stderr, "\n selpg --l lines    : lines/page")
	fmt.Fprintf(os.Stderr, "\n selpg --f          : check page with '\\f'")
	fmt.Fprintf(os.Stderr, "\n selpg --d des      : pipe destination\n")
}

func FlagInit(args *selpg_Args) {
	//参数绑定变量
	flag.IntVar(&args.start, "s", 0, "the start Page")             //开始页码，默认为0
	flag.IntVar(&args.end, "e", 0, "the end Page")                 //结束页码，默认为0
	flag.IntVar(&args.lines, "l", 72, "the lines of the page")     //每页行数，默认为72行每页
	flag.StringVar(&args.des, "d", "", "the destination to print") //输出位置，默认为空字符

	//分析指令并设置参数
	//读取方式（l按行数计算页，f按换页符计算页）
	//查找 f
	isF := flag.Bool("f", false, "")
	flag.Parse()

	//如果输入f，按照f并取-1；否则按照 l
	if *isF {
		args.readType = "f"
		args.lines = -1
	} else {
		args.readType = "l"
	}

	//输入方式（文件输入还是键盘输入）
	//如果使用了文件输入，将方式置为文件名
	args.inputType = ""
	if flag.NArg() == 1 {
		args.inputType = flag.Arg(0)
	}

	//判断参数是否合法
	//检查剩余参数数量
	if narg := flag.NArg(); narg != 1 && flag.NArg() != 0 {
		usage()
		os.Exit(1)
	}
	//检查起始终止页
	if args.start > args.end || args.start < 1 {
		usage()
		os.Exit(1)
	}
	//检查l f 是否同时出现
	if args.readType == "f" && args.lines != -1 {
		usage()
		os.Exit(1)
	}
}

//执行指令
//判断输入方式，并将输入流绑定-》如果有管道，绑定管道-》l/f读取
func run(args *selpg_Args) {
	//初始化
	fin := os.Stdin           //输入
	fout := os.Stdout         //输出
	currentLine := 0          //当前行
	currentPage := 1          //当前页
	var inpipe io.WriteCloser //管道
	var err error             //错误

	//判断输入方式
	if args.inputType != "" {
		fin, err = os.Open(args.inputType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Can't find input file \"%s\"!\n", args.inputType)
			usage()
			os.Exit(1)
		}
		defer fin.Close() //全部结束了再关闭
	}

	//确定输出到文件或者输出到屏幕
	//通过用管道接通grep模拟打印机测试，结果输出到屏幕
	if args.des != "" {
		cmd := exec.Command("grep", "-nf", "keyword")
		inpipe, err = cmd.StdinPipe()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer inpipe.Close() //最后执行
		cmd.Stdout = fout
		cmd.Start()
	}

	//分页方式
	//设置行数
	if args.readType == "l" {
		//按照行读取
		line := bufio.NewScanner(fin)
		for line.Scan() {
			if currentPage >= args.start && currentPage <= args.end {
				//输出到窗口
				fout.Write([]byte(line.Text() + "\n"))
				if args.des != "" {
					//定向到文件管道
					inpipe.Write([]byte(line.Text() + "\n"))
				}
			}
			currentLine++
			//翻页
			if currentLine == args.lines {
				currentPage++
				currentLine = 0
			}
		}
	} else {
		//用换行符 '\f'分页
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
			//'\f'翻页
			page = strings.Replace(page, "\f", "", -1)
			currentPage++
			if currentPage >= args.start && currentPage <= args.end {
				fmt.Fprintf(fout, "%s", page)
			}
		}
	}
	//当输出完成后，比较输出的页数与期望输出的数量
	if currentPage < args.end {
		fmt.Fprintf(os.Stderr, "./selpg: end_page (%d) greater than total pages (%d), less output than expected\n", args.end, currentPage)
	}
}

func main() {
	var args selpg_Args
	FlagInit(&args)
	run(&args)
}
