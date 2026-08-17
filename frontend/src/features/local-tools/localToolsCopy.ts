export interface LocalToolsCopy {
  menuLabel: string
  menuTitle: string
  ccSwitch: string
  ccSwitchDescription: string
  deepSeekHarness: string
  deepSeekHarnessDescription: string
  unavailable: string
  selectTool: string
}

export const localToolsCopy: Record<'zh' | 'en', LocalToolsCopy> = {
  zh: {
    menuLabel: '本地工具',
    menuTitle: '选择本地工具',
    ccSwitch: 'CC Switch',
    ccSwitchDescription: '导入当前 API Key 配置',
    deepSeekHarness: 'DeepSeek Harness',
    deepSeekHarnessDescription: '安装并启动本地 Harness',
    unavailable: '当前密钥不可用',
    selectTool: '选择工具'
  },
  en: {
    menuLabel: 'Local tools',
    menuTitle: 'Choose a local tool',
    ccSwitch: 'CC Switch',
    ccSwitchDescription: 'Import the current API key',
    deepSeekHarness: 'DeepSeek Harness',
    deepSeekHarnessDescription: 'Install and start the local Harness',
    unavailable: 'This key is unavailable',
    selectTool: 'Choose a tool'
  }
}
