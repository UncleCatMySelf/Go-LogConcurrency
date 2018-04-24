<h2>借用Go语言的并发优势的高性能的日志监控系统</h2>
<h2>High Performance Log monitoring system using the concurrency advantage of Go language</h2>

<p>Golang是一门简单高效的编程语言，我在编写学习的过程中也被其特性所吸引，日志监控系统是生产环境中必备的功能系统，它的作用可以说仅次于核心系统
而Golang的协程实现可以很好的帮我们完成这一核心功能，通过模拟读取nginx输出的日志文件，使用log_proccess.go进行实时读取解析写入到influxdb存储，
再由grafana进行实时展现。mock_data.go是我用于模拟日志输出的一个应用程序。</p>
<p>Golang is a simple and efficient programming language, I am in the process of writing learning is also attracted by its characteristics, log monitoring system is a necessary function of the production environment system, its role can be said after the core system
   And the Golang implementation can help us to complete this core function, through the simulation of reading Nginx output log files, using Log_proccess.go for real-time read parsing write to influxdb storage, The Grafana is then displayed in real time. Mock_data.go is an application I use to simulate log output.</p>

<p>Golang的并发实现可以通过goroutine执行，而多个goroutine间的数据同步与通信则是channel，且多个channel可以选择数据的读取与写入。
这里需要认真理解下并发与并行。并发：指同一时刻，系统通过调度，来回切换交替的运行多个任务，“看起来”是同时进行；并行：指同一时刻，
两个任务“真正的”同时进行；</p>
<p>Concurrent implementations of Golang can be performed through Goroutine, while data synchronization and communication between multiple goroutine are channel, and multiple channel can choose to read and write data. This requires a careful understanding of concurrency and parallelism.
   Concurrency: Refers to the same moment, the system through the scheduling, switching back and forth to run multiple tasks, "looks" is the same time; Two tasks "real" at the same time;</p>

1、读取模块的实现<br>
————打开文件<br>
————从文件末尾开始逐行读取<br>
————写入Read Channel<br>
2、解析模块的实现<br>
————从Read Channel中读取每行日志数据<br>
————正则提取所需的监控数据（path、status、method等）<br>
————写入Write Channel<br>
3、写入模块的实现<br>
————初始化influxdb client<br>
————从Write Channel中读取监控数据<br>
————构造数据并写入influxdb<br>
4、绘制监控图<br>
————用grafana<br>
5、监控模块的实现<br>
————总处理日志行数<br>
————系统吞出量<br>
————read channel长度<br>
————write channel长度<br>
————运行总时间<br>
————错误数<br>


<h4> Author: UncleCatMySelf</h4>
<h4>email:zhupeijie_java@126.com</h4>
<h5>what's the problem? Welcome to contact QQ:1341933031</h5>.
<h4>作者：UncleCatMySelf</h4>
<h4>email：zhupeijie_java@126.com</h4>
<h5>有什么问题，欢迎联系QQ：1341933031</h5>


