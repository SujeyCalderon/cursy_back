Write-Host "Compilando backend para Linux..."
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o main main.go

Write-Host "Subiendo binario al servidor..."
scp -i "cursyKey" main ubuntu@52.20.206.74:/home/ubuntu/cursy_back/main

Write-Host "Ajustando permisos en el servidor..."
ssh -i "cursyKey" ubuntu@52.20.206.74 "chmod +x /home/ubuntu/cursy_back/main"

Write-Host "Reiniciando servicio..."
ssh -i "cursyKey" ubuntu@52.20.206.74 "sudo systemctl restart cursy"

Write-Host "Mostrando logs..."
ssh -i "cursyKey" ubuntu@52.20.206.74 "sudo journalctl -u cursy -f"
