## Database
* a "registry" database in a global database
* some "app_gobby_id" databases which provide data storage for gobbys

### MariaDB Installation
```sh
sudo yum install mariadb mariadb-server
sudo systemctl start mariadb
sudo systemctl enable mariadb
mysql_secure_installation
```
edit '/etc/my.cnf', in section '[mysqld]', add:
```vim
init_connect='SET collation_connection = utf8_unicode_ci' 
init_connect='SET NAMES utf8' 
character-set-server=utf8 
collation-server=utf8_unicode_ci 
skip-character-set-client-handshake
```
edit '/etc/my.cnf.d/client.cnf', in section '[client]', add:
```vim
default-character-set=utf8
```
edit '/etc/my.cnf.d/mysql-clients.cnf', in section '[mysql]', add:
```vim
default-character-set=utf8
```
restart the mariadb
```sh
systemctl restart mariadb
```
validate the settings
```sh
mysql -uroot -ppassword -e "show variables like '%character%'; show variables like '%collation%';"
```

### Import Registry
```sh
mysql -uroot -ppassword -e "create database registry; show databases;"
mysql -uroot -ppassword registry < /path/to/registry.sql
mysql -uroot -ppassword -e "use registry; show tables;"
```
add some test data
```sh
mysql -uroot -ppassword -e "use registry; insert into t_global_config (t_category, t_key, t_value) values ('global', 'wechat', '1');"
mysql -uroot -ppassword -e "use registry; select * from t_global_config;"
```

### Redis
```sh
```