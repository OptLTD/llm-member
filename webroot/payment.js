// 支付页面JavaScript

class PaymentManager {
  constructor() {
    this.plans = []
    this.keyOfUsage = 'usage'
    this.selectedPlan = null;
    this.selectedPaymentMethod = null;
    this.customAmount = 0;
    this.currentOrderId = null;
    this.paymentCheckInterval = null;
  }

  updateUserDisplay(user) {
    const userNameElement = document.getElementById("userName");
    const userEmailElement = document.getElementById("userEmail");

    if (userNameElement) {
      userNameElement.textContent = user.username || user.email;
    }
    if (userEmailElement) {
      userEmailElement.textContent = user.email;
    }
  }

  bindUserMenuEvents() {
    // 用户菜单切换
    const userMenuBtn = document.getElementById("userMenuBtn");
    const userDropdown = document.getElementById("userDropdown");

    if (userMenuBtn && userDropdown) {
      userMenuBtn.addEventListener("click", (e) => {
        e.stopPropagation();
        userDropdown.classList.toggle("hidden");
      });

      // 点击其他地方关闭菜单
      document.addEventListener("click", () => {
        userDropdown.classList.add("hidden");
      });
    }

    // 退出登录
    const signOutBtn = document.getElementById("signOutBtn");
    if (signOutBtn) {
      signOutBtn.addEventListener("click", () => this.logout());
    }

    // 个人中心
    const profileBtn = document.getElementById("profileBtn");
    if (profileBtn) {
      profileBtn.addEventListener("click", () => {
        window.location.href = "/profile";
      });
    }
  }

  logout() {
    localStorage.removeItem("userToken");
    window.location.href = "/signin";
  }

  async init() {
    // bind MENU
    this.bindUserMenuEvents();
    // 加载套餐和支付方式
    await this.loadPlans();
    await this.loadPayments();

    // 自动选择从首页传递过来的套餐
    this.autoSelectPlan();
    setTimeout(() => {
      this.bindEvents();
      this.updatePayButton();
    }, 300);
  }

  async loadPlans() {
    try {
      const url = "/api/pricing-plans";
      const data = await Utils.apiRequest(url);
      if (data.plans) {
        this.plans = data.plans;
        this.renderPlans(data.plans);
      } else {
        console.error("加载套餐失败");
      }
    } catch (error) {
      console.error("加载套餐错误:", error);
    }
  }

  renderPlans(plans) {
    const plansList = document.getElementById("plansList");
    plansList.innerHTML = "";

    plans.forEach((plan) => {
      const planCard = document.createElement("div");
      planCard.className =
        "plan-card border-2 border-gray-200 rounded-lg p-4 cursor-pointer";
      planCard.dataset.plan = plan.plan;
      planCard.dataset.price = plan.price;
      planCard.dataset.name = plan.name;
      planCard.innerHTML = `
                <div class="flex items-center justify-between">
                    <div class="flex-1">
                        <div class="flex items-center">
                            <h3 class="text-lg font-semibold">${plan.name}</h3>
                            ${
                              plan.popular
                                ? '<span class="ml-2 bg-blue-500 text-white text-xs px-2 py-1 rounded-full">推荐</span>'
                                : ""
                            }
                        </div>
                        <p class="text-gray-600 text-sm mt-1">${plan.brief}</p>
                        <ul class="mt-2 text-sm text-gray-600">
                            ${plan.features
                              .map(
                                (feature) =>
                                  `<li class="flex items-center"><i class="fas fa-check text-green-500 mr-2"></i>${feature}</li>`
                              )
                              .join("")}
                        </ul>
                    </div>
                    <div class="ml-4 text-right">
                        <div class="text-2xl font-bold text-blue-600">
                            ${
                              plan.plan === this.keyOfUsage
                                ? "自定义金额" : `¥${plan.price}`
                            }
                        </div>
                        ${ plan.plan !== this.keyOfUsage
                            ? '<div class="text-sm text-gray-500">/月</div>'
                            : ""
                        }
                    </div>
                </div>
            `;

      planCard.addEventListener("click", () => this.selectPlan(plan));
      plansList.appendChild(planCard);
    });
  }

  async loadPayments() {
    try {
      const data = await Utils.apiRequest("/api/order/methods");
      if (data.methods) {
        this.renderPaymentMethods(data.methods);
      } else {
        console.error("加载支付方式失败");
      }
    } catch (error) {
      console.error("加载支付方式错误:", error);
    }
  }

  renderPaymentMethods(methods) {
    const paymentMethods = document.getElementById("paymentMethods");
    paymentMethods.innerHTML = "";

    methods.forEach((method) => {
      const methodCard = document.createElement("div");
      methodCard.className =
        "payment-method border-2 border-gray-200 rounded-lg p-3 cursor-pointer flex items-center";
      methodCard.dataset.method = method.method;

      methodCard.innerHTML = `
                <i class="${method.icon}" style="color: ${method.color}; font-size: 1.5rem;"></i>
                <span class="ml-3 font-medium">${method.name}</span>
                <i class="fas fa-check ml-auto text-gray-400"></i>
            `;

      methodCard.addEventListener("click", () =>
        this.selectPaymentMethod(method)
      );
      paymentMethods.appendChild(methodCard);
    });
  }

  autoSelectPlan() {
    // 检查localStorage中是否有预选的套餐
    const selectedPlanType = localStorage.getItem("selectedPlan");
    if (selectedPlanType) {
      // 清除localStorage中的选择，避免重复使用
      localStorage.removeItem("selectedPlan");

      // 查找对应的套餐卡片并自动选择
      setTimeout(() => {
        const planCard = document.querySelector(
          `[data-plan="${selectedPlanType}"]`
        );
        if (planCard) {
          planCard.click();
          // 滚动到套餐选择区域
          planCard.scrollIntoView({ behavior: "smooth", block: "center" });
        }
      }, 100); // 延迟确保DOM已渲染
    }
  }

  selectPlan(plan) {
    // 移除之前的选中状态
    document.querySelectorAll(".plan-card").forEach((card) => {
      card.classList.remove("selected");
    });

    // 选中当前套餐
    const planCard = document.querySelector(`[data-plan="${plan.plan}"]`);
    planCard.classList.add("selected");

    this.selectedPlan = plan;
    if (plan.plan === this.keyOfUsage) {
      // 获取基础版卡片内的金额输入框的值
      const basicAmountInput = document.querySelector("#basicAmount");
      basicAmountInput.value = plan.price;
      this.customAmount = plan.price;
    } else {
      this.customAmount = plan.price;
    }

    this.updateOrderSummary();
  }

  selectPaymentMethod(method) {
    // 移除之前的选中状态
    document.querySelectorAll(".payment-method").forEach((card) => {
      card.classList.remove("selected");
    });

    // 选中当前支付方式
    const methodCard = document.querySelector(
      `[data-method="${method.method}"]`
    );
    methodCard.classList.add("selected");

    this.selectedPaymentMethod = method;
    this.updatePayButton();
  }

  updateOrderSummary() {
    const selectedPlanName = document.getElementById("selectedPlanName");
    const selectedAmount = document.getElementById("selectedAmount");
    const basicAmountContainer = document.getElementById(
      "basicAmountContainer"
    );
    const totalAmount = document.getElementById("totalAmount");

    if (this.selectedPlan) {
      selectedPlanName.textContent = this.selectedPlan.name;

      if (this.selectedPlan.plan === this.keyOfUsage) {
        // usage方案：显示输入框，隐藏固定金额
        selectedAmount.classList.add("hidden");
        basicAmountContainer.classList.remove("hidden");
        totalAmount.textContent = `¥${this.customAmount}`;
      } else {
        // 其他方案：显示固定价格
        selectedAmount.classList.remove("hidden");
        basicAmountContainer.classList.add("hidden");
        selectedAmount.textContent = `¥${this.selectedPlan.price}`;
        totalAmount.textContent = `¥${this.selectedPlan.price}`;
      }
    } else {
      selectedPlanName.textContent = "请选择套餐";
      selectedAmount.classList.remove("hidden");
      basicAmountContainer.classList.add("hidden");
      selectedAmount.textContent = "¥0";
      totalAmount.textContent = "¥0";
    }

    this.updatePayButton();
  }

  updatePayButton() {
    const payBtn = document.getElementById("payBtn");
    if (!payBtn) return;
    
    const hasNum = (!this.selectedPlan || this.selectedPlan.plan !== this.keyOfUsage || this.customAmount > 0)
    const canPay = this.selectedPlan && this.selectedPaymentMethod && hasNum;

    payBtn.disabled = !canPay;

    // 只需要更改按钮文本和图标
    if (!canPay) {
      payBtn.innerHTML = '<i class="fas fa-lock mr-2"></i>请完成选择';
    } else {
      payBtn.innerHTML = '<i class="fas fa-credit-card mr-2"></i>立即支付';
    }
  }

  bindEvents() {
    // usage方案的金额输入框
    const basicAmountInput = document.getElementById("basicAmount");
    basicAmountInput.addEventListener("input", (e) => {
      this.customAmount = parseFloat(e.target.value) || 0;
      this.updateOrderSummary();
    });

    // 支付按钮
    const payBtn = document.getElementById("payBtn");
    payBtn.addEventListener("click", () => this.createPayment());

    // 弹窗按钮
    document
      .getElementById("checkPaymentBtn")
      .addEventListener("click", () => this.checkPaymentStatus());
    document
      .getElementById("cancelPaymentBtn")
      .addEventListener("click", () => this.cancelPayment());
    document
      .getElementById("retryPaymentBtn")
      .addEventListener("click", () => this.hidePaymentModal());
    document.getElementById("goToProfileBtn1").addEventListener("click", () => {
      window.location.href = "/profile";
    });
    document.getElementById("goToProfileBtn2").addEventListener("click", () => {
      window.location.href = "/profile";
    });
  }

  async createPayment() {
    if (!this.selectedPlan || !this.selectedPaymentMethod) {
      alert("请选择套餐和支付方式");
      return;
    }

    if (this.selectedPlan.plan === this.keyOfUsage && this.customAmount < 10) {
      alert("充值金额不能少于10元");
      return;
    }

    this.showPaymentModal();
    this.showPaymentLoading();

    try {
      const paymentData = {
        payPlan: this.selectedPlan.plan,
        amount: this.selectedPlan.price,
        method: this.selectedPaymentMethod.method,
      };

      if (this.selectedPlan.plan === this.keyOfUsage) {
        paymentData.amount = this.customAmount;
      }

      const url = "/api/order/create";
      const data = await Utils.apiRequest(url, {
        body: JSON.stringify(paymentData),
      });

      if (data && data.orderId) {
        if (!data.qrcode && data.payUrl) {
          window.open(data.payUrl, '_blank');
          this.currentOrderId = data.orderId;
          this.showPaymentRedirect();
          this.startPaymentCheck(data);
        } else {
          var qrcode = await this.loadImage(data.qrcode);
          this.currentOrderId = data.orderId;
          this.showPaymentQR(qrcode);
          this.startPaymentCheck(data);
        }
      } else if (data.error && data.plan) {
        this.showHasActivePlan(data.plan, data.expireAt);
      } else {
        this.showPaymentError(data.error || "创建支付订单失败");
      }
    } catch (error) {
      console.error("创建支付订单错误:", error);
      this.showPaymentError(error.message || "网络错误，请重试");
    }
  }

  async loadImage(url) {
    try {
      const token = Auth.token;
      const response = await fetch(url, {
        method: "POST", headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      const blob = await response.blob();
      return URL.createObjectURL(blob);
    } catch (error) {
      console.error("图片加载失败:", error);
      return "/default-image.jpg";
    }
  }

  showPaymentModal() {
    document.getElementById("paymentModal").classList.remove("hidden");
  }

  hidePaymentModal() {
    document.getElementById("paymentModal").classList.add("hidden");
    this.stopPaymentCheck();
    this.currentOrderId = null;
  }

  showPaymentLoading() {
    document.getElementById("paymentLoading").classList.remove("hidden");
    document.getElementById("paymentQR").classList.add("hidden");
    document.getElementById("paymentSuccess").classList.add("hidden");
    document.getElementById("paymentError").classList.add("hidden");
  }

  async showPaymentQR(qrcode) {
    document.getElementById("paymentLoading").classList.add("hidden");
    document.getElementById("paymentQR").classList.remove("hidden");
    document.getElementById("paymentSuccess").classList.add("hidden");
    document.getElementById("paymentError").classList.add("hidden");

    document.getElementById("qrCodeImage").src = qrcode;
    document.getElementById("paymentMethodName").textContent =
      this.selectedPaymentMethod.name;
  }

  showPaymentRedirect() {
    document.getElementById("paymentLoading").classList.add("hidden");
    document.getElementById("paymentQR").classList.add("hidden");
    document.getElementById("paymentSuccess").classList.add("hidden");
    document.getElementById("paymentError").classList.add("hidden");
    document.getElementById("hasActivePlan").classList.add("hidden");
    
    // 移除已存在的跳转提示
    const existingRedirect = document.getElementById("paymentRedirect");
    if (existingRedirect) {
      existingRedirect.remove();
    }
    
    // 显示跳转支付提示信息
    const paymentContainer = document.querySelector("#paymentModal .text-center");
    if (paymentContainer) {
      const redirectInfo = document.createElement("div");
      redirectInfo.id = "paymentRedirect";
      redirectInfo.innerHTML = `
        <div class="mb-4">
          <i class="fas fa-external-link-alt text-blue-500 text-4xl mb-2"></i>
        </div>
        <h3 class="text-lg font-semibold mb-2">已跳转到支付页面</h3>
        <p class="text-gray-600 mb-4">请在新打开的页面中完成支付</p>
        <p class="text-sm text-gray-500 mb-4">支付完成后，此页面将自动更新</p>
        <button id="cancelPaymentBtn2" class="w-full bg-gray-500 text-white py-2 rounded-lg hover:bg-gray-600">
          取消支付
        </button>
      `;
      
      paymentContainer.appendChild(redirectInfo);
      redirectInfo.classList.remove("hidden");
      
      // 绑定取消按钮事件
      const cancelBtn = document.getElementById("cancelPaymentBtn2");
      if (cancelBtn) {
        cancelBtn.addEventListener("click", () => {
          this.cancelPayment();
        });
      }
    }
  }

  showPaymentSuccess() {
    document.getElementById("paymentLoading").classList.add("hidden");
    document.getElementById("paymentQR").classList.add("hidden");
    document.getElementById("paymentSuccess").classList.remove("hidden");
    document.getElementById("paymentError").classList.add("hidden");
    
    // 移除跳转支付的提示信息
    const existingRedirect = document.getElementById("paymentRedirect");
    if (existingRedirect) {
      existingRedirect.remove();
    }
    
    this.stopPaymentCheck();
  }

  showPaymentError(message) {
    document.getElementById("paymentLoading").classList.add("hidden");
    document.getElementById("paymentQR").classList.add("hidden");
    document.getElementById("paymentSuccess").classList.add("hidden");
    document.getElementById("paymentError").classList.remove("hidden");
    document.getElementById("errorMessage").textContent = message;
    this.stopPaymentCheck();
  }

  showHasActivePlan(name, expireAt) {
    // 隐藏其他支付相关元素
    document.getElementById("paymentLoading").classList.add("hidden");
    document.getElementById("paymentQR").classList.add("hidden");
    document.getElementById("paymentSuccess").classList.add("hidden");
    document.getElementById("paymentError").classList.add("hidden");

    // 格式化过期时间
    const expireDate = new Date(expireAt);
    const formattedDate = expireDate.toLocaleDateString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });

    
    // 填充套餐信息
    const plan = this.plans.find((p) => p.plan === name);
    document.getElementById("planName").textContent = plan?.name || "未知套餐";
    document.getElementById("expireTime").textContent = formattedDate;

    // 显示已有订单信息
    document.getElementById("hasActivePlan").classList.remove("hidden");
    this.stopPaymentCheck();
  }

  async checkPaymentStatus() {
    if (!this.currentOrderId) return;

    try {
      const data = await Utils.apiRequest(
        `/api/order/query/${this.currentOrderId}`
      );
      if (data.status === "succeed") {
        this.showPaymentSuccess();
      } else if (data.status === "failed") {
        this.showPaymentError("支付失败");
      } else if (data.status === "cancelled") {
        this.showPaymentError("支付已取消");
      }
    } catch (error) {
      console.error("检查支付状态错误:", error);
    }
  }

  startPaymentCheck(data) {
    if (data?.method == "mock") {
      return;
    }
    this.paymentCheckInterval = setInterval(() => {
      this.checkPaymentStatus();
    }, 3000); // 每3秒检查一次
  }

  stopPaymentCheck() {
    if (this.paymentCheckInterval) {
      clearInterval(this.paymentCheckInterval);
      this.paymentCheckInterval = null;
    }
  }

  cancelPayment() {
    this.hidePaymentModal();
  }
}

// 页面加载完成后初始化
document.addEventListener("DOMContentLoaded", async () => {
  const token = localStorage.getItem("userToken");
  if (!token || token.trim() === "") {
    localStorage.setItem("prevPage", "/payment");
    window.location.href = "/signin";
    return
  }

  // 检查用户登录状态并更新UI
  if (window.App && window.App.Auth) {
    await window.App.Auth.checkAuth();
    if (!window.App.Auth.user) {
      localStorage.setItem("prevPage", "/payment");
      window.location.href = "/signin";
      return
    }
    const payment = new PaymentManager();
    payment.init()
  }
});
