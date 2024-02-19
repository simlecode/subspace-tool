#nohup ./collect --start-height 118000 --mysql "admin:_Admin123@(127.0.0.1:3306)/subspace_3h_collect?parseTime=true&loc=Local" >> run-collect.log 2>&1 &

nohup ./collect --mysql "admin:_Admin123@(127.0.0.1:3306)/subspace_3h_collect?parseTime=true&loc=Local" >> run-collect.log 2>&1 &
