#!/bin/bash

# 判断系统类型
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$(echo $ID | tr '[:upper:]' '[:lower:]' | tr -d '"')
elif type lsb_release >/dev/null 2>&1; then
    OS=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
else
    echo "无法检测操作系统"
    exit 1
fi

# 根据系统类型安装MySQL
if [ "$OS" = "ubuntu" ]; then
    echo "检测到Ubuntu系统，开始安装MySQL..."
    sudo apt-get update
    sudo apt-get install -y mysql-server
elif [ "$OS" = "centos" ]; then
    echo "检测到CentOS系统，开始安装MySQL..."
    sudo yum install -y mysql-server
    sudo systemctl start mysqld
    sudo systemctl enable mysqld
else
    echo "不支持的操作系统: $OS"
    exit 1
fi

echo "MySQL安装完成！"

# 创建数据库和用户
echo "正在创建数据库和用户..."
sudo mysql -e "CREATE DATABASE IF NOT EXISTS tkmall CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
sudo mysql -e "CREATE USER 'tkmalluser'@'localhost' IDENTIFIED BY 'yourpassword';"
sudo mysql -e "GRANT ALL PRIVILEGES ON tkmall.* TO 'tkmalluser'@'localhost';"
sudo mysql -e "FLUSH PRIVILEGES;"

echo "数据库和用户创建完成！"
