# 使用MMAP写的日志
MMAP：[MMAP基本概念](https://www.cnblogs.com/huxiao-tee/p/4660352.html)

* 一、mmap可以直接将文件映射至内存，减少写文件时的拷贝次数，速度比普通写文件日志要快
```
普通日志100W次写入需要耗时：18.01s
MMAP日志100W次定稿需要耗时：1.679s
```

# 问题
因为mmap的特性，导致服务一旦crash，则日志文件会遗留大量的占位符