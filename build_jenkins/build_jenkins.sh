##制作镜像及编译
echo "$1"
docker build -f ../Dockerfile -t roulette ..

if [ $? -eq 0 ]; then ## $? 是shell 上一条命令的返回值，如果执行成功，退出码是 0，如果失败，退出码是 非0
     echo "The Job Excute Success........"
else
     echo "The Job  Excute Failed........."
     exit 1
fi

##添加镜像标签
echo "$2"
docker tag roulette harbor.rgstest.slammerstudios.com/game/roulette/roulette:latest

##推送镜像
echo "$3"
docker push harbor.rgstest.slammerstudios.com/game/roulette/roulette:latest


