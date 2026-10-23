#!/data/data/com.termux/files/usr/bin/bash
read -p "输入裸域名:" domain
mkdir -p res/$domain
subfinder -d $domain -silent > res/$domain/list.txt
nuclei -l res/$domain/list.txt -tags files > res/$domain/result.txt
echo -e "\033[32m====扫描结束====\033[0m"
cat res/$domain/result.txt

