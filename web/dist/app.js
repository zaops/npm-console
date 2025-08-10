// npm-console Web Application
class NPMConsoleApp {
    constructor() {
        this.apiBase = '/api/v1';
        this.currentSection = 'dashboard';
        this.init();
    }

    init() {
        this.setupNavigation();
        this.setupEventListeners();
        this.loadDashboard();
    }

    setupNavigation() {
        const navLinks = document.querySelectorAll('.nav-link');
        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = link.getAttribute('href').substring(1);
                this.showSection(section);
            });
        });
    }

    setupEventListeners() {
        // Refresh button
        document.getElementById('refreshBtn').addEventListener('click', () => {
            this.refreshCurrentSection();
        });

        // Clear all caches
        document.getElementById('clearAllCaches').addEventListener('click', () => {
            this.clearAllCaches();
        });

        // Search packages
        document.getElementById('searchPackages').addEventListener('click', () => {
            this.searchPackages();
        });

        // Package search on enter
        document.getElementById('packageSearch').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.searchPackages();
            }
        });

        // Install package
        document.getElementById('installPackage').addEventListener('click', () => {
            this.installPackage();
        });

        // Install package on enter
        document.getElementById('packageInstall').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.installPackage();
            }
        });


    }

    showSection(section) {
        // Hide all sections
        document.querySelectorAll('.section').forEach(s => s.classList.add('hidden'));
        
        // Show selected section
        document.getElementById(section).classList.remove('hidden');
        document.getElementById(section).classList.add('fade-in');
        
        this.currentSection = section;
        
        // Load section data
        switch (section) {
            case 'dashboard':
                this.loadDashboard();
                break;
            case 'cache':
                this.loadCacheInfo();
                break;
            case 'packages':
                this.loadPackages();
                break;
            case 'config':
                this.loadConfig();
                break;
        }
    }

    refreshCurrentSection() {
        this.showSection(this.currentSection);
    }

    async apiCall(endpoint, options = {}) {
        try {
            const response = await fetch(`${this.apiBase}${endpoint}`, {
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            if (!data.success) {
                throw new Error(data.error?.message || 'API call failed');
            }

            return data.data;
        } catch (error) {
            console.error('API call failed:', error);
            this.showToast('Error: ' + error.message, 'error');
            throw error;
        }
    }

    async loadDashboard() {
        try {
            this.showLoading();

            // Load dashboard data
            const [cacheSize, managers, globalPackages] = await Promise.all([
                this.apiCall('/cache/size'),
                this.apiCall('/managers/available'),
                this.apiCall('/packages/global')
            ]);

            // Update dashboard cards
            document.getElementById('totalCacheSize').textContent = this.formatBytes(cacheSize.total_size);
            document.getElementById('availableManagers').textContent = managers.length;
            document.getElementById('globalPackages').textContent = globalPackages.length;

            // Load manager status
            await this.loadManagerStatus();

        } catch (error) {
            console.error('Failed to load dashboard:', error);
        } finally {
            this.hideLoading();
        }
    }

    async loadManagerStatus() {
        try {
            const [configs, cacheInfos] = await Promise.all([
                this.apiCall('/config'),
                this.apiCall('/cache')
            ]);

            const statusContainer = document.getElementById('managerStatus');
            statusContainer.innerHTML = '';

            if (configs && configs.length > 0) {
                configs.forEach(config => {
                    const cacheInfo = cacheInfos ? cacheInfos.find(c => c.manager === config.manager) : null;
                    const statusCard = this.createManagerStatusCard(config, cacheInfo);
                    statusContainer.appendChild(statusCard);
                });
            } else {
                statusContainer.innerHTML = '<p class="text-gray-500 text-center py-4">暂无可用的包管理器</p>';
            }
        } catch (error) {
            console.error('Failed to load manager status:', error);
            const statusContainer = document.getElementById('managerStatus');
            statusContainer.innerHTML = '<p class="text-red-500 text-center py-4">加载管理器状态失败</p>';
        }
    }

    createManagerStatusCard(config, cacheInfo) {
        const div = document.createElement('div');
        div.className = 'flex items-center justify-between p-4 border border-gray-200 rounded-lg';

        const cacheSize = cacheInfo ? this.formatBytes(cacheInfo.size) : '未知';
        const registry = config.registry || '未设置';
        const proxy = config.proxy || '无';

        div.innerHTML = `
            <div class="flex items-center space-x-4">
                <div class="flex-shrink-0">
                    <i class="fas fa-cube text-blue-500 text-xl"></i>
                </div>
                <div>
                    <h4 class="text-sm font-medium text-gray-900 uppercase">${config.manager}</h4>
                    <p class="text-sm text-gray-500">Cache: ${cacheSize}</p>
                </div>
            </div>
            <div class="text-right">
                <p class="text-xs text-gray-500">镜像源: ${this.truncateUrl(registry)}</p>
                <p class="text-xs text-gray-500">代理: ${proxy === '无' ? proxy : this.truncateUrl(proxy)}</p>
            </div>
        `;

        return div;
    }

    async loadCacheInfo() {
        try {
            this.showLoading();
            const cacheInfos = await this.apiCall('/cache');

            const cacheList = document.getElementById('cacheList');
            cacheList.innerHTML = '';

            cacheInfos.forEach(cache => {
                const cacheCard = this.createCacheCard(cache);
                cacheList.appendChild(cacheCard);
            });
        } catch (error) {
            console.error('Failed to load cache info:', error);
        } finally {
            this.hideLoading();
        }
    }

    createCacheCard(cache) {
        const div = document.createElement('div');
        div.className = 'flex items-center justify-between p-4 border border-gray-200 rounded-lg';

        div.innerHTML = `
            <div class="flex items-center space-x-4">
                <div class="flex-shrink-0">
                    <i class="fas fa-database text-blue-500 text-xl"></i>
                </div>
                <div>
                    <h4 class="text-sm font-medium text-gray-900 uppercase">${cache.manager}</h4>
                    <p class="text-sm text-gray-500">${cache.path}</p>
                    <p class="text-xs text-gray-400">${cache.file_count} files</p>
                </div>
            </div>
            <div class="flex items-center space-x-4">
                <div class="text-right">
                    <p class="text-sm font-medium text-gray-900">${this.formatBytes(cache.size)}</p>
                    <p class="text-xs text-gray-500">${this.formatDate(cache.last_updated)}</p>
                </div>
                <button onclick="app.clearCache('${cache.manager}')" 
                        class="bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded text-xs transition-colors">
                    Clear
                </button>
            </div>
        `;

        return div;
    }

    async loadPackages() {
        try {
            this.showLoading();
            const packages = await this.apiCall('/packages/global');

            const packageList = document.getElementById('packageList');
            packageList.innerHTML = '';

            if (packages.length === 0) {
                packageList.innerHTML = '<p class="text-gray-500 text-center py-8">未找到全局包</p>';
                return;
            }

            packages.slice(0, 50).forEach(pkg => { // Limit to first 50 packages
                const packageCard = this.createPackageCard(pkg);
                packageList.appendChild(packageCard);
            });

            if (packages.length > 50) {
                const moreDiv = document.createElement('div');
                moreDiv.className = 'text-center py-4';
                moreDiv.innerHTML = `<p class="text-gray-500">显示 ${packages.length} 个包中的 50 个</p>`;
                packageList.appendChild(moreDiv);
            }
        } catch (error) {
            console.error('Failed to load packages:', error);
        } finally {
            this.hideLoading();
        }
    }

    createPackageCard(pkg) {
        const div = document.createElement('div');
        div.className = 'flex items-center justify-between p-4 border border-gray-200 rounded-lg';

        div.innerHTML = `
            <div class="flex items-center space-x-4">
                <div class="flex-shrink-0">
                    <i class="fas fa-box text-purple-500 text-xl"></i>
                </div>
                <div>
                    <h4 class="text-sm font-medium text-gray-900">${pkg.name}</h4>
                    <p class="text-sm text-gray-500">${pkg.description || '无描述'}</p>
                </div>
            </div>
            <div class="flex items-center space-x-4">
                <div class="text-right">
                    <p class="text-sm font-medium text-gray-900">${pkg.version}</p>
                    <p class="text-xs text-gray-500 uppercase">${pkg.manager}</p>
                </div>
                <button onclick="app.uninstallPackage('${pkg.name}', '${pkg.manager}')"
                        class="bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded text-xs transition-colors">
                    <i class="fas fa-trash mr-1"></i> 卸载
                </button>
            </div>
        `;

        return div;
    }

    async loadConfig() {
        try {
            this.showLoading();
            const configs = await this.apiCall('/config');

            const configList = document.getElementById('configList');
            configList.innerHTML = '';

            configs.forEach(config => {
                const configCard = this.createConfigCard(config);
                configList.appendChild(configCard);
            });
        } catch (error) {
            console.error('Failed to load config:', error);
        } finally {
            this.hideLoading();
        }
    }

    createConfigCard(config) {
        const div = document.createElement('div');
        div.className = 'border border-gray-200 rounded-lg p-6';

        div.innerHTML = `
            <div class="flex items-center justify-between mb-4">
                <h4 class="text-lg font-medium text-gray-900 uppercase">${config.manager}</h4>
                <i class="fas fa-cog text-gray-400"></i>
            </div>
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">镜像源</label>
                    <div class="flex space-x-2">
                        <input type="url" value="${config.registry || ''}" 
                               class="flex-1 border border-gray-300 rounded-md px-3 py-2 text-sm"
                               id="registry-${config.manager}">
                        <button onclick="app.updateRegistry('${config.manager}')"
                                class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm transition-colors">
                            更新
                        </button>
                    </div>
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">代理</label>
                    <div class="flex space-x-2">
                        <input type="url" value="${config.proxy || ''}" 
                               class="flex-1 border border-gray-300 rounded-md px-3 py-2 text-sm"
                               id="proxy-${config.manager}">
                        <button onclick="app.updateProxy('${config.manager}')"
                                class="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-md text-sm transition-colors">
                            更新
                        </button>
                    </div>
                </div>
            </div>
        `;

        return div;
    }





    // Action methods
    async clearAllCaches() {
        if (!confirm('确定要清空所有缓存吗？此操作无法撤销。')) {
            return;
        }

        try {
            this.showLoading();
            await this.apiCall('/cache', { method: 'DELETE' });
            this.showToast('所有缓存已成功清空', 'success');
            this.loadCacheInfo();
        } catch (error) {
            console.error('Failed to clear caches:', error);
        } finally {
            this.hideLoading();
        }
    }

    async clearCache(manager) {
        if (!confirm(`确定要清空 ${manager} 缓存吗？`)) {
            return;
        }

        try {
            await this.apiCall(`/cache/${manager}`, { method: 'DELETE' });
            this.showToast(`${manager} 缓存已成功清空`, 'success');
            this.loadCacheInfo();
        } catch (error) {
            console.error(`Failed to clear ${manager} cache:`, error);
        }
    }

    async searchPackages() {
        const query = document.getElementById('packageSearch').value.trim();
        if (!query) {
            this.loadPackages();
            return;
        }

        try {
            this.showLoading();
            const packages = await this.apiCall(`/packages/search?q=${encodeURIComponent(query)}`);

            const packageList = document.getElementById('packageList');
            packageList.innerHTML = '';

            if (packages.length === 0) {
                packageList.innerHTML = '<p class="text-gray-500 text-center py-8">未找到包</p>';
                return;
            }

            packages.forEach(pkg => {
                const packageCard = this.createPackageCard(pkg);
                packageList.appendChild(packageCard);
            });
        } catch (error) {
            console.error('Failed to search packages:', error);
        } finally {
            this.hideLoading();
        }
    }

    async updateRegistry(manager) {
        const registryInput = document.getElementById(`registry-${manager}`);
        const registry = registryInput.value.trim();

        if (!registry) {
            this.showToast('镜像源URL是必需的', 'error');
            return;
        }

        try {
            await this.apiCall(`/config/${manager}/registry`, {
                method: 'PUT',
                body: JSON.stringify({ registry })
            });
            this.showToast(`${manager} 的镜像源已更新`, 'success');
        } catch (error) {
            console.error(`Failed to update registry for ${manager}:`, error);
        }
    }

    async updateProxy(manager) {
        const proxyInput = document.getElementById(`proxy-${manager}`);
        const proxy = proxyInput.value.trim();

        try {
            if (proxy) {
                await this.apiCall(`/config/${manager}/proxy`, {
                    method: 'PUT',
                    body: JSON.stringify({ proxy })
                });
                this.showToast(`${manager} 的代理已更新`, 'success');
            } else {
                await this.apiCall(`/config/${manager}/proxy`, { method: 'DELETE' });
                this.showToast(`${manager} 的代理已移除`, 'success');
            }
        } catch (error) {
            console.error(`Failed to update proxy for ${manager}:`, error);
        }
    }





    // Utility methods
    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    }

    formatDate(dateString) {
        if (!dateString || dateString === '0001-01-01T00:00:00Z') return '从未';
        return new Date(dateString).toLocaleDateString('zh-CN');
    }

    truncateUrl(url, maxLength = 30) {
        if (url.length <= maxLength) return url;
        return url.substring(0, maxLength) + '...';
    }

    showLoading() {
        document.getElementById('loadingOverlay').classList.remove('hidden');
    }

    hideLoading() {
        document.getElementById('loadingOverlay').classList.add('hidden');
    }

    showToast(message, type = 'info') {
        const toast = document.createElement('div');
        const bgColor = type === 'success' ? 'bg-green-500' : type === 'error' ? 'bg-red-500' : 'bg-blue-500';
        
        toast.className = `${bgColor} text-white px-6 py-3 rounded-lg shadow-lg flex items-center space-x-2 fade-in`;
        toast.innerHTML = `
            <i class="fas fa-${type === 'success' ? 'check' : type === 'error' ? 'exclamation-triangle' : 'info'}-circle"></i>
            <span>${message}</span>
        `;

        document.getElementById('toastContainer').appendChild(toast);

        setTimeout(() => {
            toast.remove();
        }, 5000);
    }

    async installPackage() {
        const packageName = document.getElementById('packageInstall').value.trim();
        const manager = document.getElementById('managerSelect').value;

        if (!packageName) {
            this.showToast('请输入包名', 'error');
            return;
        }

        if (!confirm(`确定要使用 ${manager} 安装 ${packageName} 吗？`)) {
            return;
        }

        try {
            this.showLoading(true);
            await this.apiCall(`/packages/install`, {
                method: 'POST',
                body: JSON.stringify({
                    name: packageName,
                    manager: manager,
                    global: true
                })
            });

            this.showToast(`${packageName} 安装成功`, 'success');
            document.getElementById('packageInstall').value = '';

            // 刷新包列表
            this.loadPackages();
            this.loadDashboard();
        } catch (error) {
            console.error('Failed to install package:', error);
            this.showToast(`安装 ${packageName} 失败: ${error.message}`, 'error');
        } finally {
            this.showLoading(false);
        }
    }

    async uninstallPackage(packageName, manager) {
        if (!confirm(`确定要卸载 ${packageName} 吗？此操作无法撤销。`)) {
            return;
        }

        try {
            this.showLoading(true);
            await this.apiCall(`/packages/uninstall`, {
                method: 'POST',
                body: JSON.stringify({
                    name: packageName,
                    manager: manager,
                    global: true
                })
            });

            this.showToast(`${packageName} 卸载成功`, 'success');

            // 刷新包列表
            this.loadPackages();
            this.loadDashboard();
        } catch (error) {
            console.error('Failed to uninstall package:', error);
            this.showToast(`卸载 ${packageName} 失败: ${error.message}`, 'error');
        } finally {
            this.showLoading(false);
        }
    }
}

// Initialize the application
const app = new NPMConsoleApp();
