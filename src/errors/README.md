### 自定义错误实现

通过实现Error()string方法, 即实现golang标准error接口, 从而方便用户自定义错误对象;

golang中错误表示目前有如下三种方法:

#### 1.golang标准包error 
  
  适合简单error message

#### 2.通过fmt.Errorf()方法可以进一步丰富error message信息

这种方式适合临时富态错误信息定义; 缺点是不便于统一维护(如公共自定义统一修订, 统一初始化等) 

#### 3.通过自定义Error()string方法体 实现error接口, 从而实现自定义错误类型;

这种方式非常灵活, 可根据业务需要自行设计, 如需要支持自动插入时间, 文本行号, 支持自定义error code 与error message 关联, 甚至支持error template 复用及error prefix等;

> 通常较为复杂的项目, 尤其是微服务设计时, 各个微服务间采用统一的error设计非常重要, 一套设计合理的自定义错误结构和方法, 不仅有利于统一维护, 更便于快速错误定位及排查; 
