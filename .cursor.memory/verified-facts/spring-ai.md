# Spring AI 已核实事实（扩充资料时引用，避免幻觉）

> 最后核对：2026-06-30
> 官方：https://docs.spring.io/spring-ai/reference/

## 版本

- Spring Boot 3.2.x / 3.3.x 与 Spring AI 1.0.x 搭配（BOM 管理）
- JDK 17+

## 依赖（OpenAI 兼容 / DeepSeek）

```xml
<dependencyManagement>
  <dependencies>
    <dependency>
      <groupId>org.springframework.ai</groupId>
      <artifactId>spring-ai-bom</artifactId>
      <version>1.0.0</version>
      <type>pom</type>
      <scope>import</scope>
    </dependency>
  </dependencies>
</dependencyManagement>
<dependency>
  <groupId>org.springframework.ai</groupId>
  <artifactId>spring-ai-openai-spring-boot-starter</artifactId>
</dependency>
```

## DeepSeek 配置（OpenAI 兼容）

```yaml
spring:
  ai:
    openai:
      api-key: ${DEEPSEEK_API_KEY}
      base-url: https://api.deepseek.com
      chat:
        options:
          model: deepseek-chat
```

## Ollama

```yaml
spring:
  ai:
    ollama:
      base-url: http://localhost:11434
      chat:
        options:
          model: qwen2.5:3b
```

依赖：`spring-ai-ollama-spring-boot-starter`

## ChatClient（1.0 风格）

- 注入 `ChatClient.Builder`（自动配置提供）
- 链式：`chatClient.prompt().user("...").call().content()`
- 流式：`.stream().content()` 返回 `Flux<String>`

## Tool

- `@Tool` 注解方法（需 spring-ai 支持 + 启用 tool calling）
- 或 `FunctionCallback` / `ToolCallback` 注册（以当前文档为准）

## 不确定时

资料中写：「请以你 pom 中 Spring AI 版本对应的官方文档为准」，并给官方链接。
