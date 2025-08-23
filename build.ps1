# ����Ҫ���������ƽ̨
$platforms = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Output = "AlicePushBotBurningTool-windows.exe" },
    @{ GOOS = "linux";   GOARCH = "amd64"; Output = "AlicePushBotBurningTool-linux" },
    @{ GOOS = "linux";   GOARCH = "arm64"; Output = "AlicePushBotBurningTool-arm64" },
    @{ GOOS = "darwin";  GOARCH = "amd64"; Output = "AlicePushBotBurningTool-mac" }
)

# ��¼ԭʼ��������
$originalGOOS = $env:GOOS
$originalGOARCH = $env:GOARCH

foreach ($p in $platforms) {
    # ���õ�ǰƽ̨��������
    $env:GOOS = $p.GOOS
    $env:GOARCH = $p.GOARCH
  
    # ���ɱ�������
    $cmd = "go build -o $($p.Output) ."
  
    # ���⴦����Windows ƽ̨�Զ����� .exe ��׺
    if ($p.GOOS -eq "windows") {
        $cmd = "go build -o $($p.Output) ."
    }
  
    # ִ�б���
    Write-Host "? ���ڱ��� [$($p.GOOS)-$($p.GOARCH)] ���: $($p.Output)" -Foreground Cyan
    Invoke-Expression $cmd
  
    # ����Ƿ�ɹ�
    if ($?) {
        Write-Host "? ����ɹ���" -Foreground Green
    } else {
        Write-Host "? ����ʧ�ܣ����� CGO ������ Go �汾" -Foreground Red
    }
}

# �ָ�ԭʼ��������
$env:GOOS = $originalGOOS
$env:GOARCH = $originalGOARCH

Write-Host "? ���б���������ɣ�" -Foreground Yellow
Write-Host "�����ļ��б���"
Get-ChildItem *.exe, *.app, * | Where-Object { $_.PSIsContainer -eq $false }