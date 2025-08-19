// 价格页面管理
const PricingPage = {
  // 初始化价格页面
  async init() {
    await this.loadPricingPlans();
  },

  // 从后端API获取价格信息
  async loadPricingPlans() {
    try {
      const url = '/api/pricing-plans'
      const data = await Utils.apiRequest(url);
      if (data.plans && data.plans.length > 0) {
        this.renderPricingPlans(data.plans);
      } else {
        this.showEmptyState();
      }
    } catch (error) {
      console.error("获取价格信息失败:", error);
      this.showErrorState();
    }
  },

  // 渲染价格方案
  renderPricingPlans(plans) {
    const container = document.getElementById("pricingPlans");
    if (!container) return;

    // 清空容器
    container.innerHTML = "";

    // 根据方案数量动态设置网格布局
    const planCount = plans.length;
    let gridClass = "";

    switch (planCount) {
      case 1:
        gridClass = "grid grid-cols-1 max-w-md mx-auto gap-8";
        break;
      case 2:
        gridClass = "grid grid-cols-1 md:grid-cols-2 max-w-4xl mx-auto gap-8";
        break;
      case 3:
        gridClass = "grid grid-cols-1 md:grid-cols-3 gap-8";
        break;
      case 4:
        gridClass = "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8";
        break;
      default:
        // 超过4个方案时，使用响应式网格
        gridClass =
          "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8";
    }

    container.className = gridClass;

    plans.forEach((plan, index) => {
      // 对于奇数个方案，中间的方案设为推荐；对于偶数个方案，第二个方案设为推荐
      const isRecommended =
        planCount % 2 === 1 ? index === Math.floor(planCount / 2) : index === 1;
      const planCard = this.createPlanCard(plan, isRecommended);
      container.appendChild(planCard);
    });
  },

  // 创建价格方案卡片
  createPlanCard(plan, isRecommended = false) {
    const card = document.createElement("div");
    card.className = `pricing-card bg-white rounded-xl shadow-lg p-8 border-2 ${
      isRecommended ? "border-blue-500 relative" : "border-gray-200"
    }`;

    // 推荐标签
    const recommendedBadge = isRecommended
      ? `
            <div class="absolute -top-4 left-1/2 transform -translate-x-1/2">
                <span class="bg-blue-500 text-white px-4 py-1 rounded-full text-sm font-semibold">推荐</span>
            </div>
        `
      : "";

    // 处理特性列表
    const features = Array.isArray(plan.features) ? plan.features : [];
    const featuresList = features.map((feature) => `
      <li class="flex items-center">
        <i class="fas fa-check text-green-500 mr-3"></i>
        <span>${feature}</span>
      </li>
    `).join("");

    // 价格显示
    const priceDisplay = plan.price === 0 ? "免费" : `¥${plan.price}`;
    const priceUnit = plan.price === 0 ? "" : "/月";

    // 按钮样式
    const buttonClass = isRecommended
      ? "w-full bg-blue-600 text-white py-3 rounded-lg font-semibold hover:bg-blue-700 transition duration-200"
      : "w-full bg-gray-600 text-white py-3 rounded-lg font-semibold hover:bg-gray-700 transition duration-200";

    // 获取方案的key（从名称推断）
    const planName = plan.name || "未知方案";
    const planKey = plan.plan || "basic";

    card.innerHTML = `
            ${recommendedBadge}
            <div class="h-full flex flex-col text-center">
                <div class="flex-grow">
                    <h3 class="text-2xl font-bold text-gray-900 mb-4">${planName}</h3>
                    <div class="mb-6">
                        <span class="text-4xl font-bold ${
                          isRecommended ? "text-blue-600" : "text-gray-900"
                        }">${priceDisplay}</span>
                        <span class="text-gray-600">${priceUnit}</span>
                    </div>
                    <p class="text-gray-600 mb-6">${plan.Brief || ""}</p>
                    <ul class="text-left space-y-3 mb-8">
                        ${featuresList}
                    </ul>
                </div>
                <div class="mt-auto">
                    <button class="${buttonClass}" onclick="selectPlan('${planKey}')">
                        选择${planName}
                    </button>
                </div>
            </div>
        `;

    return card;
  },

  // 显示空状态
  showEmptyState() {
    const container = document.getElementById("pricingPlans");
    if (!container) return;

    container.innerHTML = `
            <div class="col-span-full text-center py-20">
                <i class="fas fa-tags text-6xl text-gray-300 mb-4"></i>
                <h3 class="text-xl font-semibold text-gray-600 mb-2">暂无价格方案</h3>
                <p class="text-gray-500">请稍后再试或联系管理员</p>
            </div>
        `;
  },

  // 显示错误状态
  showErrorState() {
    const container = document.getElementById("pricingPlans");
    if (!container) return;

    container.innerHTML = `
            <div class="col-span-full text-center py-20">
                <i class="fas fa-exclamation-triangle text-6xl text-red-300 mb-4"></i>
                <h3 class="text-xl font-semibold text-gray-600 mb-2">加载失败</h3>
                <p class="text-gray-500 mb-4">无法获取价格信息，请检查网络连接</p>
                <button onclick="PricingPage.loadPricingPlans()" class="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition duration-200">
                    <i class="fas fa-redo mr-2"></i>重新加载
                </button>
            </div>
        `;
  },

  // 事件监听器已移动到 utils.js 中的 EventListeners.setupPricingEventListeners()
};

// 页面加载完成后初始化
document.addEventListener("DOMContentLoaded", () => {
  PricingPage.init();

  // 设置事件监听器
  if (window.App && window.App.EventListeners) {
    window.App.EventListeners.setupEventListeners();
  }

  // 检查用户登录状态并更新UI
  if (window.App && window.App.Auth) {
    window.App.Auth.checkAuth();
  }
});

// 导出到全局作用域供其他脚本使用
window.PricingPage = PricingPage;

// 将selectPlan函数添加到全局作用域
window.selectPlan = (planType) => {
  // 检查用户是否已登录
  if (!AppState.isLoggedIn) {
    localStorage.setItem("prevPage", "/payment");
    Utils.showNotification("请先登录后再选择套餐", "warning");
    window.location.href = "/signin";
    return;
  }

  // 将选择的套餐类型存储到localStorage，供支付页面使用
  localStorage.setItem("selectedPlan", planType);

  // 跳转到支付页面
  window.location.href = "/payment";
};
