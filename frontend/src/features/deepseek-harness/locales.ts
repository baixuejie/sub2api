export interface DeepSeekHarnessCopy {
  action: string
  title: string
  loading: string
  key: string
  group: string
  endpoint: string
  protocol: string
  model: string
  environment: string
  nodeRequirement: string
  helperVersion: string
  harnessVersion: string
  installAndStart: string
  cancel: string
  close: string
  retry: string
  relaunch: string
  installHelper: string
  updateHelper: string
  helperMissing: string
  helperMissingDetail: string
  openHarness: string
  completed: string
  failed: string
  expired: string
  unavailableKey: string
  stages: Record<string, string>
}

export const deepSeekHarnessCopy: Record<'zh' | 'en', DeepSeekHarnessCopy> = {
  zh: {
    action: '安装 DSH',
    title: 'DeepSeek Harness',
    loading: '正在读取配置',
    key: 'API Key',
    group: '分组',
    endpoint: '端点',
    protocol: '协议',
    model: '默认模型',
    environment: '本机环境',
    nodeRequirement: 'Node.js',
    helperVersion: 'Helper',
    harnessVersion: 'Harness',
    installAndStart: '安装并启动',
    cancel: '取消',
    close: '关闭',
    retry: '重试',
    relaunch: '重新唤起 Helper',
    installHelper: '安装本地 Helper',
    updateHelper: '更新本地 Helper',
    helperMissing: '未检测到本地 Helper',
    helperMissingDetail: '首次使用需要安装一次 Helper，安装后重新唤起即可。',
    openHarness: '打开 Harness',
    completed: 'Harness 已启动',
    failed: '安装未完成',
    expired: '安装会话已过期',
    unavailableKey: '当前密钥不可用于安装',
    stages: {
      awaiting_helper: '等待本地 Helper',
      checking_environment: '检查 Node.js 环境',
      installing: '安装 DeepSeek Harness',
      configuring: '写入模型和凭据配置',
      starting: '启动 Harness Web',
      completed: 'Harness 已启动',
      failed: '安装失败',
      expired: '会话已过期'
    }
  },
  en: {
    action: 'Install DSH',
    title: 'DeepSeek Harness',
    loading: 'Loading configuration',
    key: 'API Key',
    group: 'Group',
    endpoint: 'Endpoint',
    protocol: 'Protocol',
    model: 'Default model',
    environment: 'Local environment',
    nodeRequirement: 'Node.js',
    helperVersion: 'Helper',
    harnessVersion: 'Harness',
    installAndStart: 'Install and start',
    cancel: 'Cancel',
    close: 'Close',
    retry: 'Retry',
    relaunch: 'Launch Helper again',
    installHelper: 'Install local Helper',
    updateHelper: 'Update local Helper',
    helperMissing: 'Local Helper was not detected',
    helperMissingDetail: 'The first use requires one Helper installation. Launch it again after installation.',
    openHarness: 'Open Harness',
    completed: 'Harness is running',
    failed: 'Installation did not complete',
    expired: 'Installation session expired',
    unavailableKey: 'This API key cannot be used for installation',
    stages: {
      awaiting_helper: 'Waiting for local Helper',
      checking_environment: 'Checking Node.js environment',
      installing: 'Installing DeepSeek Harness',
      configuring: 'Writing model and credential settings',
      starting: 'Starting Harness Web',
      completed: 'Harness is running',
      failed: 'Installation failed',
      expired: 'Session expired'
    }
  }
}
