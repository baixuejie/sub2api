export default {
  batchImageGuide: {
    title: 'Batch Image Generation',
    description: 'Submit multiple prompts in one job and download the generated images when complete'
  },
  // Home Page
  home: {
    viewOnGithub: 'View on GitHub',
    viewDocs: 'View Documentation',
    docs: 'Docs',
    switchToLight: 'Switch to Light Mode',
    switchToDark: 'Switch to Dark Mode',
    dashboard: 'Dashboard',
    login: 'Login',
    getStarted: 'Get Started',
    goToDashboard: 'Go to Dashboard',
    showcase: {
      capabilities: 'Capabilities',
      engines: 'Model engines',
      eyebrow: 'AI ROUTING CORE / READY',
      titlePrefix: 'One unified API gateway for',
      titleHighlight: 'countless AI models',
      description: 'One API entry for GPT and Claude. Unified keys, resilient routing, and clear usage billing handled by Sub2API.',
      readDocs: 'Read docs',
      chooseEngine: 'Choose a model engine',
      responseReady: 'response ready',
      visualCaption: 'Request routing in progress · both engines online',
      capabilityTitle: 'One entry point for two powerful engines',
      capabilityDescription: 'Bring accounts, channels, and quotas into one clear API gateway so you can focus on the work that matters.',
      toolTitle: 'Your tools, one connected center',
      toolDescription: 'Keep Claude, Codex, OpenClaw, Hermes, DeepSeek Harness, and CC Switch workflows connected through one clear gateway.',
      toolCoreCaption: 'WORKFLOW READY',
      engineTitle: 'GPT and Claude, on demand',
      engineDescription: 'Choose the right capability for each task while keeping a stable development experience through one interface.',
      exploreModels: 'Explore model plaza',
      gptDescription: 'GPT capabilities for general questions, structured tasks, and rapid iteration.',
      claudeDescription: 'Claude capabilities for long-form work, complex reasoning, and careful writing.',
      proof: {
        gateway: 'Unified API gateway',
        failover: 'Smart failover',
        billing: 'Transparent usage billing'
      },
      status: {
        gateway: 'Gateway status',
        routing: 'Routing engine',
        reliability: 'Reliability policy',
        failover: 'Automatic failover',
        usage: 'Usage mode',
        payg: 'Pay as you go'
      },
      nav: {
        pricing: 'Pricing',
        apps: 'Apps',
        whyUs: 'Why choose us',
        steps: 'Get started in three steps',
        developers: 'Developers & teams',
        faq: 'FAQ'
      },
      pricing: {
        eyebrow: '05 / SIMPLE PRICING',
        title: 'Transparent, flexible usage pricing',
        description: 'Pay for what you actually use, without paying for idle subscriptions. Usage, balance, and quotas stay visible in one dashboard.',
        action: 'View full pricing',
        note: 'No long-term contract · Adjust quotas anytime · Pay as you go',
        features: {
          payAsYouGo: 'Pay as you go',
          transparent: 'Transparent costs',
          quotaControl: 'Quota controls'
        }
      },
      apps: {
        eyebrow: '03 / APP ECOSYSTEM',
        title: 'Your favorite apps, ready to connect',
        description: 'From desktop clients to intelligent coding tools, connect the workflows you already use through one API entry point.',
        supported: 'Supported',
        download: 'Official download',
        website: 'Official website',
        viewAll: 'View all apps',
        previewTitle: 'Tool workspace',
        previewStatus: 'Online',
        previewSelected: 'Current connection',
        previewConnected: 'Connection stable',
        previewUpdated: 'Live updates',
        previous: 'Previous apps',
        next: 'Next apps',
        items: {
          claude: {
            name: 'Claude',
            maker: 'Anthropic',
            description: 'An AI assistant for writing, analysis, and thoughtful conversations.'
          },
          codex: {
            name: 'Codex',
            maker: 'OpenAI',
            description: 'Understand code, edit projects, and complete development tasks from the terminal.'
          },
          openclaw: {
            name: 'OpenClaw',
            maker: 'Open Source',
            description: 'An extensible local AI workflow and automation tool.'
          },
          hermes: {
            name: 'Hermes',
            maker: 'Nous Research',
            description: 'An open-source agent and research tool for developers.'
          },
          deepseekHarness: {
            name: 'DeepSeek Harness',
            maker: 'DeepSeek AI',
            description: 'Bring DeepSeek capabilities into your local development environment.'
          },
          ccSwitch: {
            name: 'CC Switch',
            maker: 'Community',
            description: 'Switch between model and client configurations in seconds.'
          }
        }
      },
      whyUs: {
        eyebrow: '01 / WHY SUB2API',
        title: 'Make model access simple',
        description: 'Reliable routing, clear billing, and practical controls help individuals and teams use AI with confidence.',
        items: {
          unified: {
            title: 'One entry point, one control plane',
            description: 'Use one API key for connected models, accounts, channels, and usage in a single dashboard.'
          },
          resilient: {
            title: 'Smart routing that stays reliable',
            description: 'Route requests by model and channel health, then fail over smoothly when a limit is reached.'
          },
          transparent: {
            title: 'Every usage cost is clear',
            description: 'See requests, tokens, and cost details in real time, with quotas and budgets for every key.'
          },
          private: {
            title: 'Your data and permissions stay yours',
            description: 'Granular access, audit records, and self-hosting support the security needs of real projects.'
          },
          reliability: {
            title: 'A reliable request path',
            description: 'Multi-channel routing and automatic failover keep critical requests moving.'
          },
          control: {
            title: 'Complete usage control',
            description: 'Set quotas by member, project, or API key and know exactly how the budget is being used.'
          },
          pricing: {
            title: 'Transparent usage billing',
            description: 'Review every charge clearly and pay only for the tokens you actually use.'
          },
          support: {
            title: 'Fits your existing workflow',
            description: 'Use familiar SDKs and AI apps without rebuilding the projects you already have.'
          }
        }
      },
      steps: {
        eyebrow: '02 / GET STARTED',
        title: 'Go live in three steps',
        description: 'From sign-up to your first request in just a few minutes.',
        items: {
          account: {
            number: '01',
            title: 'Create an account',
            description: 'Sign up, open the console, and create your first API key.'
          },
          configure: {
            number: '02',
            title: 'Configure your workflow',
            description: 'Add the unified Base URL and API key to your SDK, client, or automation tool.'
          },
          launch: {
            number: '03',
            title: 'Launch and iterate',
            description: 'Make your first request, then improve routing and quotas with logs and usage data.'
          }
        }
      },
      recommendedModels: {
        eyebrow: '04 / RECOMMENDED MODELS',
        title: 'Choose the right model for every task',
        description: 'GPT and Claude cover everyday questions, coding, long-form work, and complex reasoning.',
        gpt: {
          name: 'GPT',
          label: 'General & development',
          description: 'Fast, mature, and versatile for code, automation, and structured tasks.',
          action: 'Explore GPT models'
        },
        claude: {
          name: 'Claude',
          label: 'Reasoning & creative work',
          description: 'Long context and nuanced writing for research, analysis, and creative work.',
          action: 'Explore Claude models'
        }
      },
      developerEnterprise: {
        eyebrow: '06 / BUILT FOR TEAMS',
        title: 'From personal projects to production teams',
        description: 'One clear API foundation helps developers iterate quickly while giving teams the visibility and governance they need.',
        developer: {
          title: 'Freedom for developers',
          description: 'Use familiar OpenAI and Anthropic SDKs and connect existing projects by changing one endpoint.',
          points: ['OpenAI / Anthropic API compatible', 'Clear request and usage logs', 'Consistent local and production setup'],
          action: 'Read developer docs'
        },
        enterprise: {
          title: 'Control for teams',
          description: 'Manage access and budgets by member, project, and API key while keeping clear boundaries as you scale.',
          points: ['Team members and role permissions', 'Quotas, budgets, and audit records', 'Multi-channel routing and failover'],
          action: 'Talk to our team'
        }
      },
      faq: {
        eyebrow: '07 / FAQ',
        title: 'Still have questions?',
        description: 'Here are answers to the questions people ask before getting started.',
        items: {
          what: {
            question: 'What is Sub2API?',
            answer: 'Sub2API is a unified AI API gateway that brings models, accounts, and channels into one stable, manageable entry point.'
          },
          access: {
            question: 'How do I get started?',
            answer: 'Create an account, generate an API key, and point your app or client to the unified Base URL.'
          },
          models: {
            question: 'Which models are supported?',
            answer: 'GPT and Claude are the current focus of the home experience. Check the model list in your console for available channels.'
          },
          pricing: {
            question: 'How does pricing work?',
            answer: 'Billing is based on actual API usage, with balance, detailed usage data, and quota controls. See the pricing page for details.'
          },
          security: {
            question: 'Is my data secure?',
            answer: 'Permissions, quotas, and logs help you control access. With self-hosting, your data and credentials remain in your own infrastructure.'
          },
          setup: {
            question: 'Do I need to change my existing code?',
            answer: 'Usually you only need to point your SDK Base URL and API key to Sub2API, while keeping the rest of your code unchanged.'
          },
          support: {
            question: 'Where can I get help?',
            answer: 'Start with the docs and console logs. For further help, use the contact details provided by the project or your deployment.'
          }
        }
      }
    },
    footerTagline: 'Unified AI API gateway',
    // User-focused value proposition
    heroSubtitle: 'One Key, All AI Models',
    heroDescription: 'No need to manage multiple subscriptions. Access Claude and GPT with a single API key',
    tags: {
      subscriptionToApi: 'Subscription to API',
      stickySession: 'Session Persistence',
      realtimeBilling: 'Pay As You Go'
    },
    // Pain points section
    painPoints: {
      title: 'Sound Familiar?',
      items: {
        expensive: {
          title: 'High Subscription Costs',
          desc: 'Paying for multiple AI subscriptions that add up every month'
        },
        complex: {
          title: 'Account Chaos',
          desc: 'Managing scattered accounts and API keys across different platforms'
        },
        unstable: {
          title: 'Service Interruptions',
          desc: 'Single accounts hitting rate limits and disrupting your workflow'
        },
        noControl: {
          title: 'No Usage Control',
          desc: "Can't track where your money goes or limit team member usage"
        }
      }
    },
    // Solutions section
    solutions: {
      title: 'We Solve These Problems',
      subtitle: 'Three simple steps to stress-free AI access'
    },
    features: {
      unifiedGateway: 'One-Click Access',
      unifiedGatewayDesc: 'Get a single API key to call all connected AI models. No separate applications needed.',
      multiAccount: 'Always Reliable',
      multiAccountDesc: 'Smart routing across multiple upstream accounts with automatic failover. Say goodbye to errors.',
      balanceQuota: 'Pay What You Use',
      balanceQuotaDesc: 'Usage-based billing with quota limits. Full visibility into team consumption.'
    },
    // Comparison section
    comparison: {
      title: 'Why Choose Us?',
      headers: {
        feature: 'Comparison',
        official: 'Official Subscriptions',
        us: 'Our Platform'
      },
      items: {
        pricing: {
          feature: 'Pricing',
          official: 'Fixed monthly fee, pay even if unused',
          us: 'Pay only for what you use'
        },
        models: {
          feature: 'Model Selection',
          official: 'Single provider only',
          us: 'Switch between models freely'
        },
        management: {
          feature: 'Account Management',
          official: 'Manage each service separately',
          us: 'Unified key, one dashboard'
        },
        stability: {
          feature: 'Stability',
          official: 'Single account rate limits',
          us: 'Multi-account pool, auto-failover'
        },
        control: {
          feature: 'Usage Control',
          official: 'Not available',
          us: 'Quotas & detailed analytics'
        }
      }
    },
    providers: {
      title: 'Supported AI Models',
      description: 'One API, Multiple Choices',
      supported: 'Supported',
      soon: 'Soon',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      more: 'More'
    },
    // CTA section
    cta: {
      title: 'Ready to Get Started?',
      description: 'Sign up now and get free trial credits to experience seamless AI access',
      button: 'Sign Up Free'
    },
    footer: {
      allRightsReserved: 'All rights reserved.'
    }
  },

  // Key Usage Query Page
  keyUsage: {
    title: 'API Key Usage',
    subtitle: 'Enter your API Key to view real-time spending and usage status',
    placeholder: 'sk-ant-mirror-xxxxxxxxxxxx',
    query: 'Query',
    querying: 'Querying...',
    privacyNote: 'Your Key is processed locally in the browser and will not be stored',
    dateRange: 'Date Range:',
    dateRangeToday: 'Today',
    dateRange7d: '7 Days',
    dateRange30d: '30 Days',
    dateRange90d: '90 Days',
    dateRangeCustom: 'Custom',
    apply: 'Apply',
    used: 'Used',
    detailInfo: 'Detail Information',
    tokenStats: 'Token Statistics',
    dailyDetail: 'Daily Detail',
    modelStats: 'Model Usage Statistics',
    // Table headers
    date: 'Date',
    model: 'Model',
    requests: 'Requests',
    inputTokens: 'Input Tokens',
    outputTokens: 'Output Tokens',
    cacheCreationTokens: 'Cache Creation',
    cacheReadTokens: 'Cache Read',
    cacheWriteTokens: 'Cache Write',
    totalTokens: 'Total Tokens',
    cost: 'Cost',
    // Status
    quotaMode: 'Key Quota Mode',
    walletBalance: 'Wallet Balance',
    // Ring card titles
    totalQuota: 'Total Quota',
    limit5h: '5-Hour Limit',
    limitDaily: 'Daily Limit',
    limit7d: '7-Day Limit',
    limitWeekly: 'Weekly Limit',
    limitMonthly: 'Monthly Limit',
    // Detail rows
    remainingQuota: 'Remaining Quota',
    expiresAt: 'Expires At',
    todayExpires: '(expires today)',
    daysLeft: '({days} days)',
    usedQuota: 'Used Quota',
    resetNow: 'Resetting soon',
    subscriptionType: 'Subscription Type',
    subscriptionExpires: 'Subscription Expires',
    // Usage stat cells
    todayRequests: 'Today Requests',
    todayInputTokens: 'Today Input',
    todayOutputTokens: 'Today Output',
    todayTokens: 'Today Tokens',
    todayCacheCreation: 'Today Cache Creation',
    todayCacheRead: 'Today Cache Read',
    todayCost: 'Today Cost',
    rpmTpm: 'RPM / TPM',
    totalRequests: 'Total Requests',
    totalInputTokens: 'Total Input',
    totalOutputTokens: 'Total Output',
    totalTokensLabel: 'Total Tokens',
    totalCacheCreation: 'Total Cache Creation',
    totalCacheRead: 'Total Cache Read',
    totalCost: 'Total Cost',
    avgDuration: 'Avg Duration',
    // Messages
    enterApiKey: 'Please enter an API Key',
    querySuccess: 'Query successful',
    queryFailed: 'Query failed',
    queryFailedRetry: 'Query failed, please try again later',
    noDailyUsage: 'No daily usage data',
  },

  // Setup Wizard
  setup: {
    title: 'Sub2API Setup',
    description: 'Configure your Sub2API instance',
    database: {
      title: 'Database Configuration',
      description: 'Connect to your PostgreSQL database',
      host: 'Host',
      port: 'Port',
      username: 'Username',
      password: 'Password',
      databaseName: 'Database Name',
      sslMode: 'SSL Mode',
      passwordPlaceholder: 'Password',
      ssl: {
        disable: 'Disable',
        require: 'Require',
        verifyCa: 'Verify CA',
        verifyFull: 'Verify Full'
      }
    },
    redis: {
      title: 'Redis Configuration',
      description: 'Connect to your Redis server',
      host: 'Host',
      port: 'Port',
      username: 'Username (optional)',
      password: 'Password (optional)',
      database: 'Database',
      usernamePlaceholder: 'Leave empty for default user',
      passwordPlaceholder: 'Password',
      enableTls: 'Enable TLS',
      enableTlsHint: 'Use TLS when connecting to Redis (public CA certs)'
    },
    admin: {
      title: 'Admin Account',
      description: 'Create your administrator account',
      email: 'Email',
      password: 'Password',
      confirmPassword: 'Confirm Password',
      passwordPlaceholder: 'Min 8 characters',
      confirmPasswordPlaceholder: 'Confirm password',
      passwordMismatch: 'Passwords do not match'
    },
    ready: {
      title: 'Ready to Install',
      description: 'Review your configuration and complete setup',
      database: 'Database',
      redis: 'Redis',
      adminEmail: 'Admin Email'
    },
    status: {
      testing: 'Testing...',
      success: 'Connection Successful',
      testConnection: 'Test Connection',
      installing: 'Installing...',
      completeInstallation: 'Complete Installation',
      completed: 'Installation completed!',
      redirecting: 'Redirecting to login page...',
      restarting: 'Service is restarting, please wait...',
      timeout: 'Service restart is taking longer than expected. Please refresh the page manually.'
    }
  },

  // Common
}
