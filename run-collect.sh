#nohup ./collect --start-height 118000 --mysql "admin:_Admin123@(127.0.0.1:3306)/subspace_3h_collect?parseTime=true&loc=Local" >> run-collect.log 2>&1 &

nohup ./collect --start-height 140000 --mysql "admin:_Admin123@(127.0.0.1:3306)/subspace_3h_collect?parseTime=true&loc=Local" >> run-collect2.log 2>&1 &
