package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"modfinal/model"
	"strconv"
	"syscall"
)

/*
停止容器
1. 找到进程pid杀掉进程
2. 改变配置文件中容器的状态为stop
 */
func Stop(containerName string,tty bool)  {
	containerInfo,_:= model.GetContainerInfo(containerName)

	if containerInfo.Pid==""{
		log.Println("container not exist!")
		return
	}
	pid,err:=strconv.Atoi(containerInfo.Pid)
	if err!=nil{
		log.Fatal("stop.go ",err)
	}
	if !tty{	// 后台运行的程序才杀进程
		if err:=syscall.Kill(pid,syscall.SIGTERM);err!=nil{
			//log.Fatal("stop.go kill ERROR,",err)
			log.Println("stop.go kill ERROR,",err)
			// 这里进程太容易出错了,先不Fatel
		}
	}

	containerInfo.Status= model.STOP
	containerInfo.Pid=""
	UpdateContainerInfo(containerInfo)
	log.Println("成功停止容器")
}

func UpdateContainerInfo(containerInfo *model.ContainerInfo){
	jsonInfo,_:=json.Marshal(containerInfo)
	location:=fmt.Sprintf(model.INFOLOCATION,containerInfo.Name)
	file:=location+"/"+ model.CONFIGNAME
	if err:=ioutil.WriteFile(file,[]byte(jsonInfo),0622);err!=nil{
		log.Fatal("更新容器信息失败,",err)
	}
}