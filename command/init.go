package command

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

/*
Init 初始化容器,主要是挂载文件系统,然后运行cmd,替换当前进程为要执行的程序进程
*/
//func Init(command string) {
func Init(){

	command:=readFromPipe()

	// TODO: 注意这里
	// https://github.com/xianlubird/mydocker/issues/41#issuecomment-478799767
	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显示
	//声明你要这个新的mount namespace独立。
	// syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		log.Fatal("init.go22,", err)
		return
	}

	pwd,err:=os.Getwd()
	if err!=nil{
		log.Fatal("init.go os.Getwd(),",err)
	}
	log.Println("当前工作目录:",pwd)
	// 改变root
	pivotRoot(pwd)

	// MS_NOEXEC 本文件系统不允许执行其他程序
	// MS_NOSUID 不允许 set-user-ID 和 set-group-ID
	// MS_NODEV  默认参数
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Fatal("init.go 444 ", err)
		return
	}
	// cmd:=exec.Command(command)
	// cmd.Stdin=os.Stdin
	// cmd.Stdout=os.Stdout
	// cmd.Stderr=os.Stderr
	// if err=cmd.Run();err!=nil{
	// 	log.Fatal("init.go1",err)
	// }

	log.Println("开始运行指定程序:",command)
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		log.Fatal("init.go333 ", err.Error())
	}
}

/*
从管道中读取命令
uintptr(3)表示序号为3的文件描述符,本来有0,1,2三个,然后父进程传入了一个reader管道读端文件描述符
 */
func readFromPipe() string{
	reader:=os.NewFile(uintptr(3),"pipe")
	command,err:=ioutil.ReadAll(reader)
	if err!=nil{
		log.Fatal("init.go 从管道读数据失败,",err)
	}
	log.Println("Init 读取命令:", string(command))
	return string(command)
}

/*
使用pivot_root实现根目录的转变
 */
func pivotRoot(root string) error {
	if err:=syscall.Mount(root,root,"bind",syscall.MS_BIND|syscall.MS_REC,"");err!=nil{
		log.Fatal("init.go pivot Mount ERROR",err)
	}

	pivotDir:=filepath.Join(root,".pivot_root")

	if err:=os.Mkdir(pivotDir,0777);err!=nil{
		log.Fatal("init.go pivot Mkdir ERROR",err)
	}

	if err:=syscall.PivotRoot(root,pivotDir);err!=nil{
		log.Fatal("init.go pivot PivotRoot ERROR",err)
	}

	if err:=syscall.Chdir("/");err!=nil{
		log.Fatal("init.go pivot Chdir ERROR",err)
	}

	pivotDir=filepath.Join("/",".pivot_root")

	if err:=syscall.Unmount(pivotDir,syscall.MNT_DETACH);err!=nil{
		log.Fatal("init.go pivot Unmount ERROR",err)
	}
	log.Println("改变当前目录为根目录成功")
	return os.Remove(pivotDir)
}
