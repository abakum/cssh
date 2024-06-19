## cssh

#Как собрать:

`git clone https://github.com/abakum/cssh`
`cd cssh`
`go install`

#Для чего:
Чтоб запускать как socks5 прокси
 - Например прокси от https://www.vpnjantit.com/free-ssh
 - создай в `~/.ssh/config` алиас `Host cssh`
```
Host cssh
 User foo-vpnjantit.com
 HostName bar.vpnjantit.com
 SessionType none
 DynamicForward 127.0.0.1:1080
 PubkeyAuthentication no
 UserKnownHostsFile ~/.ssh/bar
 LogLevel debug1
```
 - запусти `cssh -password=123`
- видишь `encPassword ....`
- допиши `encPassword ....` после `Host cssh`
- в самый верх `~/.ssh/config` пиши `IgnoreUnknown *` чтоб ssh не ругался
- И что этот `~/.ssh/config` если попадёт к другому раскодирует `encPassword ....` в `123`? - Да раскодирует!

- Поэтому переименуй `cssh` например в `secret` (никому не говори) и переименуй алиас в `Host secret`
- запусти `secret -password=123`
- замени `encPassword ....` после  `Host secret`
- потом просто запускай `secret`