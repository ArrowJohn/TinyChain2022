CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

scp -i TinyChain.pem POW_ghostProtocol ubuntu@ec2-18-183-135-240.ap-northeast-1.compute.amazonaws.com:~/.

ssh -i TinyChain.pem ubuntu@ec2-18-183-135-240.ap-northeast-1.compute.amazonaws.com "cd POW/; " 
