// 全局状态管理
const ProfileApp = {
  user: null,
  token: localStorage.getItem("userToken"),
  currentTab: "usage",
  apiKey: null,
  usage: null,
};

// 数据管理
const DataManager = {
  // 加载用户资料
  async loadUserProfile() {
    try {

      const data = await Utils.apiRequest(
        `/api/profile`,{ method: "GET" }
      );
      ProfileApp.user = data.user;
      this.updateProfileUI();
    } catch (error) {
      Utils.showNotification(error.message, "error");
      // 如果未授权，跳转到登录页
      if (error.message.includes("未授权")) {
        window.location.href = "/";
      }
    }
  },

  // 更新用户资料
  async updateUserProfile(profileData) {
    try {
      const data = await Utils.apiRequest(`/api/profile`, {
        method: "PUT", body: JSON.stringify(profileData),
      });
      ProfileApp.user = data.user;
      this.updateProfileUI();
      Utils.showNotification("个人资料更新成功", "success");
    } catch (error) {
      Utils.showNotification(error.message, "error");
    }
  },

  // 加载使用统计
  async loadUsageStats() {
    try {
      const data = await Utils.apiRequest(`/api/usage`);
      ProfileApp.usage = data;
      this.updateUsageUI();
    } catch (error) {
      Utils.showNotification(error.message, "error");
    }
  },

  // 加载API密钥
  async loadAPIKeys() {
    try {
      const data = await Utils.apiRequest(`/api/api-keys`);
      ProfileApp.apiKey = data.api_key;
      this.updateAPIKeysUI();
    } catch (error) {
      Utils.showNotification(error.message, "error");
    }
  },

  // 重新生成API密钥
  async regenerateAPIKey() {
    if (!confirm("确定要重新生成API密钥吗？旧密钥将立即失效。")) {
      return;
    }

    try {
      const data = await Utils.apiRequest(`/api/regenerate`, {
        method: "POST",
      });
      ProfileApp.apiKey = data.api_key;
      this.updateAPIKeysUI();
      Utils.showNotification("API密钥重新生成成功", "success");
    } catch (error) {
      Utils.showNotification(error.message, "error");
    }
  },

  // 加载定价方案
  async loadPricingPlans() {
    const upgradeOptions = document.getElementById("upgradeOptions");
    if (!upgradeOptions) {
      return;
    }
    try {
      const data = await Utils.apiRequest(`/api/pricing-plans`);
      this.updateBillingUI(upgradeOptions, data.plans);
    } catch (error) {
      Utils.showNotification(error.message, "error");
    }
  },

  // 订阅套餐
  async subscribePlan(planId) {
    try {
      const data = await Utils.apiRequest(`/api/subscribe`, {
        body: JSON.stringify({ plan_id: planId }),
      });
      ProfileApp.user = data.user;
      this.updateProfileUI();
      Utils.showNotification("套餐订阅成功", "success");
    } catch (error) {
      Utils.showNotification(error.message, "error");
    }
  },

  // 加载订单历史
  async loadOrderHistory() {
    try {
      const data = await Utils.apiRequest(`/api/orders`);
      this.updateOrderHistoryUI(data.orders);
    } catch (error) {
      Utils.showNotification(error.message, "error");
    }
  },

  // 更新个人资料UI
  updateProfileUI() {
    if (!ProfileApp.user) return;

    const user = ProfileApp.user;

    // 更新导航栏用户信息
    const userNameEl = document.getElementById("userName");
    if (userNameEl) {
      userNameEl.textContent = user.username;
    }

    // 更新侧边栏用户信息
    const sidebarUserNameEl = document.getElementById("sidebarUserName");
    if (sidebarUserNameEl) {
      sidebarUserNameEl.textContent = user.username;
    }

    const sidebarUserEmailEl = document.getElementById("sidebarUserEmail");
    if (sidebarUserEmailEl) {
      sidebarUserEmailEl.textContent = user.email;
    }

    const userPlanEl = document.getElementById("userPlan");
    if (userPlanEl) {
      userPlanEl.textContent = this.getPlanName(user.pay_plan);
    }

    // 更新模态框用户信息
    const modalUserNameEl = document.getElementById("modalUserName");
    if (modalUserNameEl) {
      modalUserNameEl.textContent = user.username;
    }

    const modalUserEmailEl = document.getElementById("modalUserEmail");
    if (modalUserEmailEl) {
      modalUserEmailEl.textContent = user.email;
    }

    const modalUserPlanEl = document.getElementById("modalUserPlan");
    if (modalUserPlanEl) {
      modalUserPlanEl.textContent = this.getPlanName(user.pay_plan);
    }

    // 更新模态框表单字段
    const modalProfileUsernameEl = document.getElementById(
      "modalProfileUsername"
    );
    if (modalProfileUsernameEl) {
      modalProfileUsernameEl.value = user.username;
    }

    const modalProfileEmailEl = document.getElementById("modalProfileEmail");
    if (modalProfileEmailEl) {
      modalProfileEmailEl.value = user.email;
    }

    const modalProfileCreatedAtEl = document.getElementById(
      "modalProfileCreatedAt"
    );
    if (modalProfileCreatedAtEl) {
      modalProfileCreatedAtEl.value = Utils.formatDate(user.created_at);
    }

    const modalProfileStatusEl = document.getElementById("modalProfileStatus");
    if (modalProfileStatusEl) {
      modalProfileStatusEl.textContent = user.is_active ? "正常" : "已禁用";
      modalProfileStatusEl.className = user.is_active
        ? "inline-block bg-green-100 text-green-800 text-sm px-3 py-1 rounded-full"
        : "inline-block bg-red-100 text-red-800 text-sm px-3 py-1 rounded-full";
    }

    // 更新原有表单字段（如果存在）
    const profileUsernameEl = document.getElementById("profileUsername");
    if (profileUsernameEl) {
      profileUsernameEl.value = user.username;
    }

    const profileEmailEl = document.getElementById("profileEmail");
    if (profileEmailEl) {
      profileEmailEl.value = user.email;
    }

    const profileCreatedAtEl = document.getElementById("profileCreatedAt");
    if (profileCreatedAtEl) {
      profileCreatedAtEl.value = Utils.formatDate(user.created_at);
    }

    const profileStatusEl = document.getElementById("profileStatus");
    if (profileStatusEl) {
      profileStatusEl.textContent = user.is_active ? "正常" : "已禁用";
      profileStatusEl.className = user.is_active
        ? "inline-block bg-green-100 text-green-800 text-sm px-3 py-1 rounded-full"
        : "inline-block bg-red-100 text-red-800 text-sm px-3 py-1 rounded-full";
    }
  },

  // 更新使用统计UI
  updateUsageUI() {
    if (!ProfileApp.usage) return;

    const usage = ProfileApp.usage;

    document.getElementById("totalRequests").textContent = Utils.formatNumber(
      usage.total_requests
    );
    document.getElementById("totalTokens").textContent = Utils.formatNumber(
      usage.total_tokens
    );
    document.getElementById("currentPlan").textContent =
      this.getPlanName(usage.current_plan);

    // 更新进度条（这里需要实际的当日/当月使用数据）
    const dailyUsed = 0; // 需要从API获取
    const monthlyUsed = usage.total_requests; // 简化处理

    document.getElementById(
      "dailyUsage"
    ).textContent = `${dailyUsed} / ${Utils.formatNumber(usage.daily_limit)}`;
    document.getElementById("monthlyUsage").textContent = `${Utils.formatNumber(
      monthlyUsed
    )} / ${Utils.formatNumber(usage.monthly_limit)}`;

    const dailyPercent = (dailyUsed / usage.daily_limit) * 100;
    const monthlyPercent = (monthlyUsed / usage.monthly_limit) * 100;

    document.getElementById("dailyProgress").style.width = `${Math.min(
      dailyPercent,
      100
    )}%`;
    document.getElementById("monthlyProgress").style.width = `${Math.min(
      monthlyPercent,
      100
    )}%`;
  },

  // 更新API密钥UI
  updateAPIKeysUI() {
    if (!ProfileApp.apiKey) return;

    const apiKeyDisplay = document.getElementById("apiKeyDisplay");
    apiKeyDisplay.value = ProfileApp.apiKey;
  },

  // 更新套餐管理UI
  updateBillingUI(upgradeOptions, plans) {
    if (!plans || !ProfileApp.user) return;

    const currentPlan = ProfileApp.user.pay_plan;
    const currentPlanData = plans.find(
      (p) => p.plan === currentPlan
    );

    if (currentPlanData) {
      document.getElementById("currentPlanName").textContent =
        currentPlanData.name;
      document.getElementById("currentPlanDesc").textContent =
        currentPlanData.features[0];
    }
    upgradeOptions.innerHTML = "";
    plans.forEach((plan) => {
      if (p.plan !== currentPlan) {
        const planCard = this.createPlanCard(plan);
        upgradeOptions.appendChild(planCard);
      }
    });
  },

  // 更新订单历史UI
  updateOrderHistoryUI(orders) {
    const orderHistory = document.getElementById("orderHistory");
    orderHistory.innerHTML = "";

    if (!orders || orders.length === 0) {
      orderHistory.innerHTML = `
          <tr>
              <td colspan="5" class="px-6 py-4 text-center text-gray-500">暂无订单记录</td>
          </tr>
      `;
      return;
    }

    orders.forEach((order) => {
      const row = document.createElement("tr");
      row.className = "bg-white border-b hover:bg-gray-50";

      const statusClass = this.getOrderStatusClass(order.status);
      const statusText = this.getOrderStatusText(order.status);

      row.innerHTML = `
                <td class="px-6 py-4 font-medium text-gray-900">${
                  order.order_id
                }</td>
                <td class="px-6 py-4 text-gray-900">${this.getPlanName(
                  order.pay_plan
                )}</td>
                <td class="px-6 py-4 text-gray-900">¥${order.amount}</td>
                <td class="px-6 py-4">
                    <span class="px-2 py-1 text-xs font-medium rounded-full ${statusClass}">
                        ${statusText}
                    </span>
                </td>
                <td class="px-6 py-4 text-gray-500">${Utils.formatDate(
                  order.succeed_at
                )}</td>
            `;

      orderHistory.appendChild(row);
    });
  },

  // 获取订单状态样式类
  getOrderStatusClass(status) {
    const statusClasses = {
      pending: "bg-yellow-100 text-yellow-800",
      succeed: "bg-green-100 text-green-800",
      failed: "bg-red-100 text-red-800",
      refunded: "bg-red-100 text-gray-800",
      canceled: "bg-gray-100 text-gray-800",
    };
    return statusClasses[status] || "bg-gray-100 text-gray-800";
  },

  // 获取订单状态文本
  getOrderStatusText(status) {
    const statusTexts = {
      pending: "待支付",
      succeed: "已支付",
      failed: "支付失败",
      canceled: "已取消",
      refunded: "已退款",
    };
    return statusTexts[status] || status;
  },

  // 创建套餐卡片
  createPlanCard(plan) {
    const card = document.createElement("div");
    card.className =
      "border border-gray-200 rounded-lg p-6 hover:border-blue-500 transition duration-200";
    card.innerHTML = `
            <div class="text-center">
                <h4 class="text-lg font-semibold text-gray-900 mb-2">${
                  plan.name
                }</h4>
                <div class="mb-4">
                    <span class="text-2xl font-bold text-gray-900">¥${
                      plan.price
                    }</span>
                    <span class="text-gray-600">/月</span>
                </div>
                <ul class="text-sm text-gray-600 mb-6 space-y-1">
                    ${plan.features
                      .map((feature) => `<li>• ${feature}</li>`)
                      .join("")}
                </ul>
                <button class="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 transition duration-200" 
                        onclick="DataManager.subscribePlan('${plan.id}')">
                    升级到${plan.name}
                </button>
            </div>
        `;
    return card;
  },

  // 获取套餐显示名称
  getPlanName(plan) {
    const planNames = {
      basic: "基础版",
      extra: "基础版",
      ultra: "专业版",
      super: "企业版",
    };
    return planNames[plan] || plan;
  },
};

// 标签页管理
const TabManager = {
  // 切换标签页
  switchTab(tabName) {
    // 隐藏所有标签页内容
    document.querySelectorAll(".tab-content").forEach((tab) => {
      tab.classList.add("hidden");
    });

    // 移除所有标签按钮的激活状态
    document.querySelectorAll(".tab-btn").forEach((btn) => {
      btn.classList.remove("tab-active");
    });

    // 显示当前标签页
    document.getElementById(`${tabName}-tab`).classList.remove("hidden");
    document
      .querySelector(`[data-tab="${tabName}"]`)
      .classList.add("tab-active");

    ProfileApp.currentTab = tabName;

    // 加载对应的数据
    this.loadTabData(tabName);
  },

  // 加载标签页数据
  loadTabData(tabName) {
    switch (tabName) {
      case "profile":
        // 个人资料数据已在初始化时加载
        break;
      case "usage":
        DataManager.loadUsageStats();
        break;
      case "api-keys":
        DataManager.loadAPIKeys();
        break;
      case "billing":
        DataManager.loadPricingPlans();
        DataManager.loadOrderHistory();
        break;
    }
  },
};

// 事件监听器已移动到 utils.js 中的 EventListeners.setupProfileEventListeners()

// 初始化检查
async function checkAuth() {
  if (!ProfileApp.token) {
    // 显示登录提示并跳转到登录页面
    Utils.showNotification("请先登录", "warning");
    setTimeout(() => {
      window.location.href = "/signin";
    }, 1500);
    return false;
  }

  // 验证token有效性
  try {
    await Utils.apiRequest(
      `/api/profile`, { method: "GET" }
    );
    return true;
  } catch (error) {
    // token无效，已在apiRequest中处理跳转
    return false;
  }
}

// 页面加载完成后初始化
document.addEventListener("DOMContentLoaded", async () => {
  const isAuthenticated = await checkAuth();
  if (!isAuthenticated) return;

  // 设置事件监听器
  if (window.App && window.App.EventListeners) {
    window.App.EventListeners.setupEventListeners();
  }

  // 标签页切换
  document.querySelectorAll(".tab-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      const tabName = btn.getAttribute("data-tab");
      TabManager.switchTab(tabName);
    });
  });

  DataManager.loadUserProfile();

  // 默认显示使用统计标签页
  TabManager.switchTab("usage");

  console.log("个人中心页面已初始化");
});

// 导出全局对象供调试使用
window.ProfileApp = {
  ProfileApp,
  DataManager,
  TabManager,
  Utils,
};
