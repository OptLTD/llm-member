// Dashboard管理模块
class DashboardManager {
  constructor(app) {
    this.app = app
  }

  async loadDashboard() {
    try {
      const resp = await this.app.apiCall(`/api/admin/stats`);
      if (resp) {
        this.updateStats(resp);
        this.updateCharts(resp);
      }
    } catch (error) {
      console.error("Error loading dashboard data:", error);
    }
  }

  updateStats(data) {
    // API统计
    document.getElementById("totalRequests").textContent =
      data.totalRequests || 0;
    document.getElementById("successRate").textContent = `${(
      data.successRate || 0
    ).toFixed(1)}%`;
    document.getElementById("totalTokens").textContent = (
      data.totalTokens || 0
    ).toLocaleString();
    document.getElementById("inputTokens").textContent = (
      data.inputTokens || 0
    ).toLocaleString();
    document.getElementById("outputTokens").textContent = (
      data.outputTokens || 0
    ).toLocaleString();

    // 会员统计
    document.getElementById("totalMembers").textContent =
      data.totalMembers || 0;
    document.getElementById("paidMembers").textContent = data.paidMembers || 0;
    document.getElementById("totalMembersCount").textContent =
      data.totalMembers || 0;
    document.getElementById("monthlyNewMembers").textContent =
      data.monthlyNewMembers || 0;
    document.getElementById("monthlyNewPaidMembers").textContent =
      data.monthlyNewPaidMembers || 0;
    document.getElementById("monthlyNewMembersCount").textContent =
      data.monthlyNewMembers || 0;

    // 订单统计
    document.getElementById("totalOrders").textContent = data.totalOrders || 0;
    document.getElementById("successfulOrders").textContent =
      data.successfulOrders || 0;
    document.getElementById("totalOrdersCount").textContent =
      data.totalOrders || 0;
    document.getElementById("totalRevenue").textContent = `¥${(
      data.totalRevenue || 0
    ).toLocaleString()}`;
    document.getElementById("successfulRevenue").textContent = `¥${(
      data.successfulRevenue || 0
    ).toLocaleString()}`;
    document.getElementById("totalRevenueAmount").textContent = `¥${(
      data.totalRevenue || 0
    ).toLocaleString()}`;
    document.getElementById("monthlyRevenue").textContent = `¥${(
      data.monthlyRevenue || 0
    ).toLocaleString()}`;
    document.getElementById("monthlySuccessfulRevenue").textContent = `¥${(
      data.monthlySuccessfulRevenue || 0
    ).toLocaleString()}`;
    document.getElementById("monthlyTotalRevenue").textContent = `¥${(
      data.monthlyRevenue || 0
    ).toLocaleString()}`;
    document.getElementById("weeklyRevenue").textContent = `¥${(
      data.weeklyRevenue || 0
    ).toLocaleString()}`;
    document.getElementById("weeklySuccessfulRevenue").textContent = `¥${(
      data.weeklySuccessfulRevenue || 0
    ).toLocaleString()}`;
    document.getElementById("weeklyTotalRevenue").textContent = `¥${(
      data.weeklyRevenue || 0
    ).toLocaleString()}`;
  }

  // 更新图表显示
  updateCharts(data) {
    // 模型使用统计
    const modelChart = document.getElementById("modelUsageChart");
    modelChart.innerHTML = "";

    if (data.modelUsage) {
      const modelEntries = Object.entries(data.modelUsage);
      if (modelEntries.length > 0) {
        modelEntries.forEach(([model, count]) => {
          const percentage = ((count / data.totalRequests) * 100).toFixed(1);
          const bar = document.createElement("div");
          bar.className = "flex items-center mb-2";
          bar.innerHTML = `
            <div class="w-20 text-sm text-gray-600">${model}</div>
            <div class="flex-1 mx-2">
                <div class="bg-gray-200 rounded-full h-2">
                    <div class="bg-blue-500 h-2 rounded-full" style="width: ${percentage}%"></div>
                </div>
            </div>
            <div class="w-16 text-sm text-gray-600 text-right">${count}</div>
          `;
          modelChart.appendChild(bar);
        });
      } else {
        modelChart.innerHTML =
          '<p class="text-gray-500 text-center">暂无数据</p>';
      }
    }

    // 提供商使用统计
    const providerChart = document.getElementById("providerUsageChart");
    providerChart.innerHTML = "";

    if (data.providerUsage) {
      const providerEntries = Object.entries(data.providerUsage);
      if (providerEntries.length > 0) {
        providerEntries.forEach(([provider, count]) => {
          const percentage = ((count / data.totalRequests) * 100).toFixed(1);
          const bar = document.createElement("div");
          bar.className = "flex items-center mb-2";
          bar.innerHTML = `
            <div class="w-20 text-sm text-gray-600">${provider}</div>
            <div class="flex-1 mx-2">
                <div class="bg-gray-200 rounded-full h-2">
                    <div class="bg-green-500 h-2 rounded-full" style="width: ${percentage}%"></div>
                </div>
            </div>
            <div class="w-16 text-sm text-gray-600 text-right">${count}</div>
          `;
          providerChart.appendChild(bar);
        });
      } else {
        providerChart.innerHTML =
          '<p class="text-gray-500 text-center">暂无数据</p>';
      }
    }
  }
}