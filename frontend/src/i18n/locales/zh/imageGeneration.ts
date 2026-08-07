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
      clear: '清空结果',
      download: '下载图片'
    },
    results: {
      title: '生成结果',
      generating: '正在生成图片...',
      generatingHint: '生成时间取决于上游模型和图片数量。',
      empty: '输入提示词并点击生成，结果会显示在这里。',
      imageAlt: '生成的图片 {index}'
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
      generate: '图片生成失败，请稍后重试。'
    }
  }
}
