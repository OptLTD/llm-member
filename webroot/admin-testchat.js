// 测试聊天管理模块
class TestChatManager {
  constructor(app) {
    this.app = app;
    this.bindEvents();
  }

  // 绑定事件
  bindEvents() {
    // 发送消息按钮
    const sendBtn = document.getElementById("sendMessage");
    if (sendBtn) {
      sendBtn.addEventListener("click", () => this.sendChatMessage());
    }

    // 回车发送消息
    const messageInput = document.getElementById("messageInput");
    if (messageInput) {
      messageInput.addEventListener("keypress", (e) => {
        if (e.key === "Enter" && !e.shiftKey) {
          e.preventDefault();
          this.sendChatMessage();
        }
      });
    }
  }

  // 加载模型列表
  async loadModels() {
    try {
      const data = await this.app.apiCall("/api/admin/models");
      this.updateModelSelect(data.data || []);
    } catch (error) {
      if (error.message !== "Unauthorized") {
        console.error("Failed to load models:", error);
      }
    }
  }

  // 更新模型选择器
  updateModelSelect(models) {
    const select = document.getElementById("modelSelect");
    select.innerHTML = '<option value="">请选择模型</option>';

    // 获取上次选择的模型
    const lastSelectedModel = localStorage.getItem("lastSelectedModel");

    models.forEach((model) => {
      const option = document.createElement("option");
      option.value = model.id;
      option.textContent = `${model.name} (${model.provider})`;

      // 如果是上次选择的模型，设为选中状态
      if (model.id === lastSelectedModel) {
        option.selected = true;
      }

      select.appendChild(option);
    });

    // 添加模型选择变化监听器，保存选择
    select.addEventListener("change", (e) => {
      if (e.target.value) {
        localStorage.setItem("lastSelectedModel", e.target.value);
      }
    });
  }

  // 发送聊天消息
  async sendChatMessage() {
    const model = document.getElementById("modelSelect").value;
    const message = document.getElementById("messageInput").value.trim();
    const temperature = parseFloat(
      document.getElementById("temperature").value
    );
    const maxTokens = parseInt(document.getElementById("maxTokens").value);
    const streamMode = document.getElementById("streamMode").checked;

    if (!model) {
      this.app.showAlert("请选择模型", "warning");
      return;
    }

    if (!message) {
      this.app.showAlert("请输入消息内容", "warning");
      return;
    }

    const sendBtn = document.getElementById("sendMessage");
    const originalText = sendBtn.innerHTML;
    sendBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-2"></i>发送中...';
    sendBtn.disabled = true;

    try {
      const startTime = Date.now();

      if (streamMode) {
        await this.sendStreamChatMessage(
          model,
          message,
          temperature,
          maxTokens,
          startTime
        );
      } else {
        const apiKey = localStorage.getItem('apiKey')
        const url = '/v1/chat/completions'
        const data = await this.app.apiCall(url, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${apiKey}`,
          },
          body: JSON.stringify({
            model: model,
            messages: [{ role: "user", content: message }],
            temperature: temperature, max_tokens: maxTokens,
          }),
        });

        const duration = Date.now() - startTime;
        this.showChatResponse(data, duration);
      }
    } catch (error) {
      if (error.message !== "Unauthorized") {
        this.showChatError("网络错误：" + error.message);
      }
    } finally {
      sendBtn.innerHTML = originalText;
      sendBtn.disabled = false;
    }
  }

  // 显示聊天响应
  showChatResponse(data, duration) {
    const responseDiv = document.getElementById("chatResponse");
    const contentDiv = document.getElementById("responseContent");
    const statsDiv = document.getElementById("responseStats");

    if (data.choices && data.choices.length > 0) {
      // 使用 pre 标签显示响应内容，保持原始格式，不重复容器的样式
      contentDiv.innerHTML = `<pre class="whitespace-pre-wrap font-mono overflow-x-auto">${this.app.escapeHtml(
        data.choices[0].message.content
      )}</pre>`;
    } else {
      contentDiv.innerHTML =
        '<pre class="whitespace-pre-wrap font-mono text-gray-500">无响应内容</pre>';
    }

    statsDiv.innerHTML = `
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
              <span class="font-medium">模型:</span> ${
                data.model || "N/A"
              }
          </div>
          <div>
              <span class="font-medium">响应时间:</span> ${duration}ms
          </div>
          <div>
              <span class="font-medium">Token 使用:</span> ${
                data.usage?.total_tokens || "N/A"
              }
          </div>
          <div>
              <span class="font-medium">完成原因:</span> ${
                data.choices?.[0]?.finish_reason || "N/A"
              }
          </div>
      </div>
    `;

    responseDiv.classList.remove("hidden");
  }

  // 显示聊天错误
  showChatError(error) {
    const responseDiv = document.getElementById("chatResponse");
    const contentDiv = document.getElementById("responseContent");
    const statsDiv = document.getElementById("responseStats");

    contentDiv.innerHTML = `<div class="text-red-500"><i class="fas fa-exclamation-triangle mr-2"></i>${error}</div>`;
    statsDiv.innerHTML = "";

    responseDiv.classList.remove("hidden");
  }

  // 发送流式聊天消息
  async sendStreamChatMessage(
    model,
    message,
    temperature,
    maxTokens,
    startTime
  ) {
    const responseDiv = document.getElementById("chatResponse");
    const contentDiv = document.getElementById("responseContent");
    const statsDiv = document.getElementById("responseStats");

    // 初始化响应区域
    responseDiv.classList.remove("hidden");
    contentDiv.innerHTML =
      '<pre class="whitespace-pre-wrap font-mono overflow-x-auto" id="streamContent"></pre>';
    statsDiv.innerHTML = `
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
              <span class="font-medium">模型:</span> ${model}
          </div>
          <div>
              <span class="font-medium">状态:</span> <span class="text-blue-600">流式接收中...</span>
          </div>
          <div>
              <span class="font-medium">Token 使用:</span> <span id="tokenCount">0</span>
          </div>
          <div>
              <span class="font-medium">完成原因:</span> <span id="finishReason">-</span>
          </div>
      </div>
    `;

    const streamContentDiv = document.getElementById("streamContent");
    let fullContent = "";
    let tokenCount = 0;

    try {
      // 获取认证token
      const apiKey = localStorage.getItem("apiKey");
      if (!apiKey) {
        throw new Error("Unauthorized");
      }

      const url = '/v1/chat/completions'
      const response = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify({
          model: model,
          messages: [{ role: "user", content: message }],
          temperature: temperature,
          max_tokens: maxTokens,
          stream: true,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `请求失败 (${response.status})`);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split("\n");

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            const data = line.slice(6);
            if (data === "[DONE]") {
              const duration = Date.now() - startTime;
              this.updateStreamStats(model, duration, tokenCount, "stop");
              return;
            }

            try {
              const parsed = JSON.parse(data);
              if (
                parsed.choices &&
                parsed.choices[0] &&
                parsed.choices[0].delta
              ) {
                const delta = parsed.choices[0].delta;
                if (delta.content) {
                  fullContent += delta.content;
                  streamContentDiv.textContent = fullContent;
                  tokenCount++;
                  document.getElementById("tokenCount").textContent =
                    tokenCount;
                }
                if (parsed.choices[0].finish_reason) {
                  document.getElementById("finishReason").textContent =
                    parsed.choices[0].finish_reason;
                }
              }
            } catch (e) {
              console.warn("Failed to parse SSE data:", data);
            }
          }
        }
      }

      const duration = Date.now() - startTime;
      this.updateStreamStats(model, duration, tokenCount, "stop");
    } catch (error) {
      this.showChatError("流式请求错误：" + error.message);
    }
  }

  // 更新流式统计信息
  updateStreamStats(model, duration, tokenCount, finishReason) {
    const statsDiv = document.getElementById("responseStats");
    statsDiv.innerHTML = `
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div>
            <span class="font-medium">模型:</span> ${model}
        </div>
        <div>
            <span class="font-medium">响应时间:</span> ${duration}ms
        </div>
        <div>
            <span class="font-medium">Token 使用:</span> ${tokenCount}
        </div>
        <div>
            <span class="font-medium">完成原因:</span> ${finishReason}
        </div>
      </div>
    `;
  }
}