export default {
  imageGeneration: {
    title: '图片创作工作台',
    description: '选择可用分组和模型，输入提示词生成图片。',
    controls: {
      title: '生成设置',
      group: '分组',
      model: '模型',
      prompt: '提示词',
      promptPlaceholder: '描述你想生成的画面、主体、风格和细节...',
      promptHint: '提示词会由服务端进行权限和内容安全校验。',
      size: '尺寸',
      customWidth: '宽度（像素）',
      customHeight: '高度（像素）',
      customSizeHint: '边长需为 {multiple} 的倍数，最大 {maxEdge}px，总像素 {minPixels} 至 {maxPixels}。',
      count: '数量',
      quality: '质量',
      format: '输出格式',
      compression: '压缩质量',
      compressionHint: '仅 JPEG 和 WebP 生效，数值越高画质越好。',
      background: '背景',
      moderation: '内容审核'
    },
    modes: {
      generate: '文生图',
      edit: '图生图'
    },
    values: {
      quality: {
        auto: '自动',
        low: '低',
        medium: '中',
        high: '高'
      },
      format: {
        png: 'PNG',
        jpeg: 'JPEG',
        webp: 'WebP'
      },
      background: {
        auto: '自动',
        opaque: '不透明',
        transparent: '透明'
      },
      moderation: {
        auto: '自动',
        low: '宽松'
      },
      size: {
        custom: '自定义尺寸'
      }
    },
    actions: {
      retry: '重试',
      generate: '生成图片',
      generating: '生成中...',
      edit: '修改图片',
      editing: '修改中...',
      clear: '清空结果',
      download: '下载图片',
      useForEdit: '用于图生图',
      settings: '设置',
      optimizePrompt: '优化提示词',
      optimizing: '优化中...',
      restorePrompt: '恢复原始提示词'
    },
    results: {
      title: '生成结果',
      generating: '正在生成图片...',
      generatingHint: '生成时间取决于上游模型和图片数量。',
      empty: '输入提示词并点击生成，结果会显示在这里。',
      imageAlt: '生成的图片 {index}',
      previewTitle: '图片预览',
      previous: '上一张',
      next: '下一张'
    },
    edit: {
      source: '参考图片',
      chooseFiles: '选择图片',
      sourceAlt: '参考图片 {index}',
      removeSource: '移除参考图',
      sourceRequired: '请先选择至少一张参考图片。',
      invalidSourceType: '仅支持 PNG、JPEG 或 WebP 图片。',
      sourceTooLarge: '单张参考图片不能超过 20 MB。',
      tooManySources: '最多选择 {max} 张参考图片。',
      sourceLoadFailed: '无法读取这张生成图片，请先下载后再上传。'
    },
    empty: {
      title: '暂无可用图片分组',
      description: '当前账号暂无可用于图片创作的分组。请联系管理员为可用分组绑定启用的 OpenAI 渠道，并在渠道中配置 gpt-image-* 模型。'
    },
    validation: {
      promptRequired: '请输入提示词。',
      promptTooLong: '提示词过长，请缩短后重试。',
      customSize: {
        positive_integer: '宽度和高度必须是正整数。',
        edge_multiple: '宽度和高度必须满足边长倍数要求。',
        max_edge: '宽度和高度不能超过模型支持的最大边长。',
        pixel_range: '总像素数不在模型支持范围内。',
        aspect_ratio: '宽高比超过模型支持范围。'
      }
    },
    errors: {
      options: '加载图片创作配置失败，请稍后重试。',
      noModel: '请选择可用的分组和模型。',
      emptyResponse: '服务端未返回可显示的图片。',
      generate: '图片生成失败，请稍后重试。',
      edit: '图片修改失败，请稍后重试。',
      optimize: '提示词优化失败，请稍后重试。',
      optimizeEmpty: '优化器返回了空提示词。'
    },
    settings: {
      title: '图片创作设置',
      description: '选择提示词优化和图片生成使用的模型及服务端 API Key。',
      back: '返回图片创作',
      promptTitle: '提示词优化',
      promptHint: '优化请求会在服务端使用所选的 OpenAI 兼容模型和 Key。',
      imageTitle: '图片生成',
      imageHint: '所选图片模型和 Key 会作为图片创作页的默认配置。',
      group: '分组',
      model: '模型',
      apiKey: '内部 API Key',
      defaultSize: '默认规格',
      defaultCount: '默认张数',
      defaultCountHint: '可选择 1 至 9 张，多张图片会以网格展示。',
      noPromptKey: '该分组没有可用的有效 Key。',
      noImageKey: '该分组没有可用于生图的有效 Key。',
      save: '保存设置',
      saving: '保存中...',
      saved: '设置已保存。',
      loadError: '加载图片创作设置失败。',
      saveError: '保存图片创作设置失败。'
    }
  }
}
