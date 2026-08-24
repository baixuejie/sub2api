export default {
  batchImageGuide: {
    title: '图片批量生成',
    description: '一次提交多条提示词，任务完成后可统一下载图片结果'
  },
  // Home Page
  home: {
    viewOnGithub: '在 GitHub 上查看',
    viewDocs: '查看文档',
    docs: '文档',
    switchToLight: '切换到浅色模式',
    switchToDark: '切换到深色模式',
    dashboard: '控制台',
    login: '登录',
    getStarted: '立即开始',
    goToDashboard: '进入控制台',
    showcase: {
      capabilities: '核心能力',
      engines: '模型引擎',
      eyebrow: 'AI ROUTING CORE / READY',
      titlePrefix: '统一API网关服务于',
      titleHighlight: '海量AI模型',
      description: '一个 API 入口，智能连接 GPT 与 Claude。统一密钥、稳定调度、清晰计费，把复杂的模型接入交给白白AI。',
      readDocs: '查看文档',
      chooseEngine: '选择模型引擎',
      responseReady: '响应已就绪',
      visualCaption: '请求正在智能路由中 · 双引擎在线',
      capabilityTitle: '一个入口，掌控两大模型引擎',
      capabilityDescription: '把账号、渠道和额度管理收拢到一个清晰的 API 网关里，专注于真正重要的工作。',
      toolTitle: '常用工具，围绕同一个中枢协作',
      toolDescription: '让 Claude、Codex、OpenClaw、Hermes、DeepSeek Harness 和 CC Switch 等工具工作流保持在同一条清晰的连接线上。',
      toolCoreCaption: 'WORKFLOW READY',
      engineTitle: 'GPT 与 Claude，按需切换',
      engineDescription: '根据任务选择合适的模型能力，并通过统一接口保持稳定的开发体验。',
      exploreModels: '探索模型广场',
      gptDescription: '适合通用问答、结构化任务与快速迭代的 GPT 模型能力。',
      claudeDescription: '适合长文本、复杂推理与高质量写作的 Claude 模型能力。',
      proof: {
        gateway: '统一 API 网关',
        failover: '智能故障切换',
        billing: '按量透明计费'
      },
      status: {
        gateway: '网关状态',
        routing: '路由引擎',
        reliability: '可靠性策略',
        failover: '自动故障切换',
        usage: '用量模式',
        payg: '按量计费'
      },
      nav: {
        pricing: '查看定价',
        apps: '常用应用',
        whyUs: '为什么选择我们',
        steps: '三步即可上线',
        developers: '开发者与企业',
        faq: '常见问题'
      },
      pricing: {
        eyebrow: '05 / SIMPLE PRICING',
        title: '透明、灵活的用量定价',
        description: '按实际调用量付费，不为闲置订阅买单。用量、余额与配额都在一个控制台里清晰可见。',
        action: '查看完整定价',
        note: '无长期合约 · 随时调整配额 · 用多少付多少',
        features: {
          payAsYouGo: '按量计费',
          transparent: '费用透明',
          quotaControl: '配额可控'
        }
      },
      apps: {
        eyebrow: '03 / APP ECOSYSTEM',
        title: '常用应用，开箱即用',
        description: '从桌面客户端到智能开发工具，使用同一个 API 入口连接你熟悉的工作流。',
        supported: '已支持',
        download: '官方下载',
        website: '官方网站',
        viewAll: '查看全部应用',
        previewTitle: '工具工作台',
        previewStatus: '在线',
        previewSelected: '当前连接',
        previewConnected: '连接稳定',
        previewUpdated: '实时更新',
        previous: '上一组应用',
        next: '下一组应用',
        items: {
          claude: {
            name: 'Claude',
            maker: 'Anthropic',
            description: '面向写作、分析与复杂对话的 AI 助手。'
          },
          codex: {
            name: 'Codex',
            maker: 'OpenAI',
            description: '在终端中理解代码、修改项目并完成开发任务。'
          },
          openclaw: {
            name: 'OpenClaw',
            maker: 'Open Source',
            description: '可扩展的本地 AI 工作流与自动化工具。'
          },
          hermes: {
            name: 'Hermes',
            maker: 'Nous Research',
            description: '面向开发者的开源智能代理与研究工具。'
          },
          deepseekHarness: {
            name: 'DeepSeek Harness',
            maker: 'DeepSeek AI',
            description: '将 DeepSeek 能力接入本地开发环境的工具。'
          },
          ccSwitch: {
            name: 'CC Switch',
            maker: 'Community',
            description: '快速切换不同模型与客户端配置。'
          }
        }
      },
      whyUs: {
        eyebrow: '01 / WHY SUB2API',
        title: '把模型接入变成一件简单的事',
        description: '稳定的路由、清晰的账单和可控的权限，让个人与团队都能放心使用 AI。',
        items: {
          unified: {
            title: '一个入口，统一管理',
            description: '用一把 API Key 访问已接入的模型，账号、渠道与用量集中在一个控制台。'
          },
          resilient: {
            title: '智能路由，更加稳定',
            description: '根据模型与渠道状态自动调度，遇到限制时平滑切换，减少请求中断。'
          },
          transparent: {
            title: '用量费用，清清楚楚',
            description: '实时查看请求、Token 与费用明细，按团队或 Key 设置配额与预算。'
          },
          private: {
            title: '数据与权限，掌握在你手里',
            description: '细粒度权限、审计记录与自托管能力，满足个人项目和生产环境的安全要求。'
          },
          reliability: {
            title: '稳定可靠的请求链路',
            description: '多渠道智能调度与自动故障切换，让关键请求保持连续。'
          },
          control: {
            title: '完整的用量控制',
            description: '按成员、项目与 API Key 设置配额，实时掌握每一笔调用。'
          },
          pricing: {
            title: '透明的按量计费',
            description: '费用明细清晰可查，只为实际使用的 Token 付费。'
          },
          support: {
            title: '兼容你的工作流',
            description: '兼容主流 SDK 与常用 AI 应用，接入现有项目无需推倒重来。'
          }
        }
      },
      steps: {
        eyebrow: '02 / GET STARTED',
        title: '三步即可上线',
        description: '从注册到第一次调用，只需要几分钟。',
        items: {
          account: {
            number: '01',
            title: '创建账号',
            description: '注册 Sub2API，进入控制台创建你的第一个 API Key。'
          },
          connect: {
            number: '01',
            title: '创建账号并连接',
            description: '注册后进入控制台，创建 API Key 并连接需要使用的模型渠道。'
          },
          configure: {
            number: '02',
            title: '配置你的工作流',
            description: '把统一 Base URL 和 API Key 填入 SDK、客户端或自动化工具。'
          },
          launch: {
            number: '03',
            title: '上线并持续迭代',
            description: '发起第一次请求，随后通过日志、配额和路由策略持续优化。'
          }
        }
      },
      recommendedModels: {
        eyebrow: '04 / RECOMMENDED MODELS',
        title: '为每一种任务选择合适的模型',
        description: 'GPT 与 Claude 覆盖日常问答、代码开发、长文本和复杂推理等场景。',
        gpt: {
          name: 'GPT',
          label: '通用与开发',
          description: '响应迅速、生态成熟，适合代码、自动化和结构化任务。',
          action: '查看 GPT 模型'
        },
        claude: {
          name: 'Claude',
          label: '推理与创作',
          description: '长上下文与细腻表达，适合研究、写作和复杂分析。',
          action: '查看 Claude 模型'
        }
      },
      developerEnterprise: {
        eyebrow: '06 / BUILT FOR TEAMS',
        title: '从个人项目到企业生产环境，都能稳定运行',
        description: '一套清晰的 API 基础设施，帮助开发者快速迭代，也让团队拥有可观测、可治理的 AI 使用方式。',
        developer: {
          title: '给开发者的自由度',
          description: '兼容主流 OpenAI 与 Anthropic SDK，统一端点即可接入现有项目。',
          points: ['OpenAI / Anthropic API 兼容', '清晰的请求与用量日志', '本地开发与生产配置一致'],
          action: '阅读开发文档'
        },
        enterprise: {
          title: '给团队的控制力',
          description: '按成员、项目与 API Key 管理权限和预算，在扩展规模的同时保持清晰边界。',
          points: ['团队成员与角色权限', '配额、预算与审计记录', '多渠道路由与故障切换'],
          action: '联系团队顾问'
        }
      },
      faq: {
        eyebrow: '07 / FAQ',
        title: '还有疑问？',
        description: '这里整理了开始使用前最常见的问题。',
        items: {
          what: {
            question: 'Sub2API 是什么？',
            answer: 'Sub2API 是一个统一的 AI API 网关，将多个模型、账号与渠道集中到一个稳定、可管理的入口。'
          },
          models: {
            question: '目前支持哪些模型？',
            answer: '当前首页重点支持 GPT 与 Claude。具体可用模型和渠道以控制台中的模型列表为准。'
          },
          pricing: {
            question: '如何收费？',
            answer: '平台按实际 API 使用量计费，并提供余额、用量明细与配额控制，具体价格请查看定价页面。'
          },
          security: {
            question: '我的数据安全吗？',
            answer: '你可以通过权限、配额和日志控制使用范围；自托管部署时，数据与凭据保留在自己的基础设施中。'
          },
          setup: {
            question: '接入需要改动现有代码吗？',
            answer: '通常只需将 SDK 的 Base URL 与 API Key 指向 Sub2API，即可继续使用现有代码。'
          },
          access: {
            question: '如何开始使用？',
            answer: '注册账号后创建 API Key，将统一 Base URL 配置到你的应用或客户端即可。'
          },
          support: {
            question: '遇到问题可以在哪里获得帮助？',
            answer: '可以先查阅文档和控制台日志；仍需帮助时，请通过项目仓库或站点提供的联系方式联系我们。'
          }
        }
      }
    },
    footerTagline: '统一 AI API 网关',
    // 新增：面向用户的价值主张
    heroSubtitle: '一个密钥，畅用多个 AI 模型',
    heroDescription: '无需管理多个订阅账号，一站式接入 Claude 与 GPT 主流 AI 服务',
    tags: {
      subscriptionToApi: '订阅转 API',
      stickySession: '会话保持',
      realtimeBilling: '按量计费'
    },
    // 用户痛点区块
    painPoints: {
      title: '你是否也遇到这些问题？',
      items: {
        expensive: {
          title: '订阅费用高',
          desc: '每个 AI 服务都要单独订阅，每月支出越来越多'
        },
        complex: {
          title: '多账号难管理',
          desc: '不同平台的账号、密钥分散各处，管理起来很麻烦'
        },
        unstable: {
          title: '服务不稳定',
          desc: '单一账号容易触发限制，影响正常使用'
        },
        noControl: {
          title: '用量无法控制',
          desc: '不知道钱花在哪了，也无法限制团队成员的使用'
        }
      }
    },
    // 解决方案区块
    solutions: {
      title: '我们帮你解决',
      subtitle: '简单三步，开始省心使用 AI'
    },
    features: {
      unifiedGateway: '一键接入',
      unifiedGatewayDesc: '获取一个 API 密钥，即可调用所有已接入的 AI 模型，无需分别申请。',
      multiAccount: '稳定可靠',
      multiAccountDesc: '智能调度多个上游账号，自动切换和负载均衡，告别频繁报错。',
      balanceQuota: '用多少付多少',
      balanceQuotaDesc: '按实际使用量计费，支持设置配额上限，团队用量一目了然。'
    },
    // 优势对比
    comparison: {
      title: '为什么选择我们？',
      headers: {
        feature: '对比项',
        official: '官方订阅',
        us: '本平台'
      },
      items: {
        pricing: {
          feature: '付费方式',
          official: '固定月费，用不完也付',
          us: '按量付费，用多少付多少'
        },
        models: {
          feature: '模型选择',
          official: '单一服务商',
          us: '多模型随意切换'
        },
        management: {
          feature: '账号管理',
          official: '每个服务单独管理',
          us: '统一密钥，一站管理'
        },
        stability: {
          feature: '服务稳定性',
          official: '单账号易触发限制',
          us: '多账号池，自动切换'
        },
        control: {
          feature: '用量控制',
          official: '无法限制',
          us: '可设配额、查明细'
        }
      }
    },
    providers: {
      title: '已支持的 AI 模型',
      description: '一个 API，多种选择',
      supported: '已支持',
      soon: '即将推出',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      more: '更多'
    },
    // CTA 区块
    cta: {
      title: '准备好开始了吗？',
      description: '注册即可获得免费试用额度，体验一站式 AI 服务',
      button: '免费注册'
    },
    footer: {
      allRightsReserved: '保留所有权利。'
    }
  },

  // Key Usage Query Page
  keyUsage: {
    title: 'API Key 用量查询',
    subtitle: '输入您的 API Key 以查看实时消费金额与使用状态',
    placeholder: 'sk-ant-mirror-xxxxxxxxxxxx',
    query: '查询',
    querying: '查询中...',
    privacyNote: '您的 Key 仅在浏览器本地处理，不会被存储',
    dateRange: '统计范围:',
    dateRangeToday: '今日',
    dateRange7d: '7 天',
    dateRange30d: '30 天',
    dateRange90d: '90 天',
    dateRangeCustom: '自定义',
    apply: '应用',
    used: '已使用',
    detailInfo: '详细信息',
    tokenStats: 'Token 统计',
    dailyDetail: '按日明细',
    modelStats: '模型用量统计',
    // Table headers
    date: '日期',
    model: '模型',
    requests: '请求数',
    inputTokens: '输入 Tokens',
    outputTokens: '输出 Tokens',
    cacheCreationTokens: '缓存创建',
    cacheReadTokens: '缓存读取',
    cacheWriteTokens: '缓存写入',
    totalTokens: '总 Tokens',
    cost: '费用',
    // Status
    quotaMode: 'Key 限额模式',
    walletBalance: '钱包余额',
    // Ring card titles
    totalQuota: '总额度',
    limit5h: '5 小时限额',
    limitDaily: '日限额',
    limit7d: '7 天限额',
    limitWeekly: '周限额',
    limitMonthly: '月限额',
    // Detail rows
    remainingQuota: '剩余额度',
    expiresAt: '过期时间',
    todayExpires: '(今日到期)',
    daysLeft: '({days} 天)',
    usedQuota: '已用额度',
    resetNow: '即将重置',
    subscriptionType: '订阅类型',
    subscriptionExpires: '订阅到期',
    // Usage stat cells
    todayRequests: '今日请求',
    todayInputTokens: '今日输入',
    todayOutputTokens: '今日输出',
    todayTokens: '今日 Tokens',
    todayCacheCreation: '今日缓存创建',
    todayCacheRead: '今日缓存读取',
    todayCost: '今日费用',
    rpmTpm: 'RPM / TPM',
    totalRequests: '累计请求',
    totalInputTokens: '累计输入',
    totalOutputTokens: '累计输出',
    totalTokensLabel: '累计 Tokens',
    totalCacheCreation: '累计缓存创建',
    totalCacheRead: '累计缓存读取',
    totalCost: '累计费用',
    avgDuration: '平均耗时',
    // Messages
    enterApiKey: '请输入 API Key',
    querySuccess: '查询成功',
    queryFailed: '查询失败',
    queryFailedRetry: '查询失败，请稍后重试',
    noDailyUsage: '暂无按日用量数据',
  },

  // Setup Wizard
  setup: {
    title: 'Sub2API 安装向导',
    description: '配置您的 Sub2API 实例',
    database: {
      title: '数据库配置',
      description: '连接到您的 PostgreSQL 数据库',
      host: '主机',
      port: '端口',
      username: '用户名',
      password: '密码',
      databaseName: '数据库名称',
      sslMode: 'SSL 模式',
      passwordPlaceholder: '密码',
      ssl: {
        disable: '禁用',
        require: '要求',
        verifyCa: '验证 CA',
        verifyFull: '完全验证'
      }
    },
    redis: {
      title: 'Redis 配置',
      description: '连接到您的 Redis 服务器',
      host: '主机',
      port: '端口',
      username: '用户名（可选）',
      password: '密码（可选）',
      database: '数据库',
      usernamePlaceholder: '默认用户留空',
      passwordPlaceholder: '密码',
      enableTls: '启用 TLS',
      enableTlsHint: '连接 Redis 时使用 TLS（公共 CA 证书）'
    },
    admin: {
      title: '管理员账户',
      description: '创建您的管理员账户',
      email: '邮箱',
      password: '密码',
      confirmPassword: '确认密码',
      passwordPlaceholder: '至少 8 个字符',
      confirmPasswordPlaceholder: '确认密码',
      passwordMismatch: '密码不匹配'
    },
    ready: {
      title: '准备安装',
      description: '检查您的配置并完成安装',
      database: '数据库',
      redis: 'Redis',
      adminEmail: '管理员邮箱'
    },
    status: {
      testing: '测试中...',
      success: '连接成功',
      testConnection: '测试连接',
      installing: '安装中...',
      completeInstallation: '完成安装',
      completed: '安装完成！',
      redirecting: '正在跳转到登录页面...',
      restarting: '服务正在重启，请稍候...',
      timeout: '服务重启时间超出预期，请手动刷新页面。'
    }
  },

  // Common
}
