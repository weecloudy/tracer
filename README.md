### Intro

依托云服务tracing平台，实现微服务(mesh) 应用性能可观测性的基础库；对接opentracing标准协议；以及新的opentelemetry标准协议(sdk->collector)；

这里使用的实现方案采用jaeger agent方式对接(client/sdk -> agent(sidecar) -> collector)；

![jarger-arch](https://raw.githubusercontent.com/cloudwego/hertz-examples/main/opentelemetry/static/jaeger-arch.png)





### Reference

1. 阿里云：https://help.aliyun.com/document_detail/90498.html (对接比较丰富)
2. 腾讯云：https://cloud.tencent.com/document/product/1463/57462 https://cloud.tencent.com/document/product/1261/62946
3. Istio: https://istio.io/latest/about/faq/#distributed-tracing
4. Jaeger: https://github.com/jaegertracing/jaeger
5. 字节 Hertz-tracer: https://github.com/hertz-contrib/tracer https://github.com/cloudwego/hertz-examples/tree/main/tracer https://github.com/cloudwego/hertz-examples/tree/main/opentelemetry
6. bilibili kratos： https://go-kratos.dev/docs/component/middleware/tracing
7. go-zero: https://go-zero.dev/cn/docs/deployment/trace
