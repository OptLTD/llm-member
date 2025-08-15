// 定价方案管理模块
class PricingManager {
  constructor(app) {
    this.app = app;
    this.bindEvents();
  }

  // 绑定事件
  bindEvents() {
    // 保存定价方案按钮
    const savePlanBtn = document.getElementById("savePlan");
    if (savePlanBtn) {
      savePlanBtn.addEventListener("click", () => this.savePricingPlan());
    }

    // 取消编辑按钮
    const cancelEditBtn = document.getElementById("cancelEdit");
    if (cancelEditBtn) {
      cancelEditBtn.addEventListener("click", () => this.hideEditPlanModal());
    }

    // 编辑付费方案弹窗
    const editPlanModal = document.getElementById("editPlanModal");
    const cancelEditPlan = document.getElementById("cancelEditPlan");
    const closePlanModalBtn = document.getElementById("closePlanModalBtn");

    if (closePlanModalBtn) {
      closePlanModalBtn.addEventListener("click", () => {
        this.hideEditPlanModal();
      });
    }

    if (cancelEditPlan) {
      cancelEditPlan.addEventListener("click", () => {
        this.hideEditPlanModal();
      });
    }

    if (editPlanModal) {
      editPlanModal.addEventListener("click", (e) => {
        if (e.target === editPlanModal) {
          this.hideEditPlanModal();
        }
      });
    }

    // 编辑付费方案表单
    const editPlanForm = document.getElementById("editPlanForm");
    if (editPlanForm) {
      editPlanForm.addEventListener("submit", (e) => {
        e.preventDefault();
        this.savePricingPlan();
      });
    }
  }

  // 加载付费方案
  async loadPricingPlans() {
    try {
      const resp = await this.app.apiCall("/api/setup/pricing");
      this.updatePricingPlansDisplay(resp?.plans || []);
    } catch (error) {
      if (error.message !== "Unauthorized") {
        console.error("Failed to load pricing plans:", error);
      }
    }
  }

  // 更新付费方案显示
  updatePricingPlansDisplay(plans) {
    const container = document.getElementById("pricing-plans-container");

    if (!plans || plans.length === 0) {
      container.innerHTML =
        '<p class="text-gray-500 text-center">暂无付费方案</p>';
      return;
    }

    const planOrder = ["basic", "extra", "ultra", "super"];
    const sortedPlans = plans.sort((a, b) => {
      return planOrder.indexOf(a.plan) - planOrder.indexOf(b.plan);
    });

    // 设置容器为横向网格布局
    container.className =
      "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6";
    container.innerHTML = sortedPlans
      .map(
        (plan) => `
            <div class="border rounded-lg p-6 ${
              plan.enabled
                ? "border-blue-200 bg-blue-50"
                : "border-gray-200 bg-gray-50"
            }">
                <div class="flex justify-between items-start mb-4">
                    <div>
                        <h3 class="text-xl font-semibold mb-2">${plan.name}</h3>
                        <div class="text-3xl font-bold text-blue-600 mb-2">
                        ¥${
                          plan.price
                        }<span class="text-sm text-gray-500">/月</span>
                        <span class="px-2 py-1 text-xs rounded-full ${
                          plan.enabled
                            ? "bg-green-100 text-green-800"
                            : "bg-gray-100 text-gray-800"
                        }">
                          ${plan.enabled ? "已启用" : "已禁用"}
                        </span>
                        </div>
                        <p class="text-gray-600">${plan.desc}</p>
                    </div>
                    <div class="flex items-center space-x-2">
                      <button onclick="app.pricingManager.editPricingPlan('${plan.plan}')"
                        class="text-blue-600 hover:text-blue-800">
                          <i class="fas fa-edit"></i>
                      </button>
                    </div>
                </div>
                <div class="mb-4">
                    <h4 class="font-medium mb-2">特性：</h4>
                    <ul class="space-y-1">
                      ${plan.features
                        .map(
                          (feature) => `
                            <li class="flex items-center text-sm">
                                <i class="fas fa-check text-green-500 mr-2"></i>
                                ${feature}
                            </li>
                      `
                        )
                        .join("")}
                    </ul>
                </div>
            </div>
        `
      )
      .join("");
  }

  // 编辑付费方案
  async editPricingPlan(type) {
    try {
      const resp = await this.app.apiCall("/api/setup/pricing");
      const plan = resp.plans.find((p) => p.plan === type);
      plan && this.showEditPlanModal(plan);
    } catch (error) {
      if (error.message !== "Unauthorized") {
        console.error("Failed to load plan for editing:", error);
      }
    }
  }

  // 显示编辑付费方案弹窗
  showEditPlanModal(plan) {
    document.getElementById("editPlanType").value = plan.plan;
    document.getElementById("editPlanName").value = plan.name;
    document.getElementById("editPlanDesc").value = plan.desc;
    document.getElementById("editPlanPrice").value = plan.price;
    document.getElementById("editPlanFeatures").value =
      plan.features.join("\n");
    document.getElementById("editPlanEnabled").checked = plan.enabled;
    document.getElementById("editPlanModal").classList.remove("hidden");
  }

  // 隐藏编辑付费方案弹窗
  hideEditPlanModal() {
    document.getElementById("editPlanModal").classList.add("hidden");
    document.getElementById("editPlanForm").reset();
  }

  // 保存付费方案
  async savePricingPlan() {
    const planType = document.getElementById("editPlanType").value;
    const planData = {
      name: document.getElementById("editPlanName").value,
      desc: document.getElementById("editPlanDesc").value,
      price: parseFloat(document.getElementById("editPlanPrice").value),
      features: document
        .getElementById("editPlanFeatures")
        .value.split("\n")
        .filter((f) => f.trim()),
      enabled: document.getElementById("editPlanEnabled").checked,
    };

    try {
      await this.app.apiCall(`/api/setup/pricing/${planType}`, {
        method: "PUT",
        body: JSON.stringify(planData),
      });

      this.app.showAlert("付费方案保存成功！", "success");
      this.hideEditPlanModal();
      this.loadPricingPlans();
    } catch (error) {
      if (error.message !== "Unauthorized") {
        this.app.showAlert(
          "保存失败: " + (error.message || "网络错误，请重试"),
          "error"
        );
      }
    }
  }
}