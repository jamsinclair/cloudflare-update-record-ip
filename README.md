# cloudflare-update-record-ip

Small Go Program that updates the IP for a Cloudflare Zone Record.

I created this early this year to keep the DNS Record for my Minecraft Server up-to-date with the IP of the current 
AWS EC2 
Instance that was running the Minecraft server. I ran the binary on start/reboot of the Instance to keep the IP
for the DNS Record up-to-date.

## Usage:
1. Update the `authEmail`, `authKey` values in `main.go`. Get your API Key from your 
['My Account' page on Cloudflare](https://www.cloudflare.com/a/account).
1. Update the `updateZoneName`, `updateRecordName` values in `main.go` with the Zone name and it's record that you
want to keep updated.
1. [Optional] If the platform you want to run this program on does not have Go available you will want to compile `main.go`
to an executable binary. 
  
  For example, to compile a binary for the default Amazon Linux EC2 Instance:
  ```bash
env GOOS=linux GOARCH=amd64 go build
  ```
  See [this blog post](https://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5) for more info on cross compiling
1. Run the binary file or `main.go`. This will automatically exit and will print whether the Record IP Update was successful or not. 