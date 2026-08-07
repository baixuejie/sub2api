export default {
  imageGeneration: {
    title: 'Image Studio',
    description: 'Choose an available group and model, then enter a prompt to generate images.',
    controls: {
      title: 'Generation settings',
      group: 'Group',
      model: 'Model',
      prompt: 'Prompt',
      promptPlaceholder: 'Describe the scene, subject, style, and details you want...',
      promptHint: 'The server validates permissions and content safety before generating.',
      size: 'Size',
      customWidth: 'Width (px)',
      customHeight: 'Height (px)',
      customSizeHint: 'Edges must be multiples of {multiple}; max {maxEdge}px; total pixels {minPixels} to {maxPixels}.',
      count: 'Quantity',
      quality: 'Quality',
      format: 'Output format',
      compression: 'Compression quality',
      compressionHint: 'Applies to JPEG and WebP only. Higher values preserve more detail.',
      background: 'Background',
      moderation: 'Moderation'
    },
    modes: {
      generate: 'Text to image',
      edit: 'Image to image'
    },
    values: {
      quality: {
        auto: 'Auto',
        low: 'Low',
        medium: 'Medium',
        high: 'High'
      },
      format: {
        png: 'PNG',
        jpeg: 'JPEG',
        webp: 'WebP'
      },
      background: {
        auto: 'Auto',
        opaque: 'Opaque',
        transparent: 'Transparent'
      },
      moderation: {
        auto: 'Auto',
        low: 'Low'
      },
      size: {
        custom: 'Custom size'
      }
    },
    actions: {
      retry: 'Retry',
      generate: 'Generate image',
      generating: 'Generating...',
      edit: 'Edit image',
      editing: 'Editing...',
      clear: 'Clear results',
      download: 'Download image',
      useForEdit: 'Use for image editing',
      settings: 'Settings',
      optimizePrompt: 'Optimize prompt',
      optimizing: 'Optimizing...',
      restorePrompt: 'Restore original prompt'
    },
    results: {
      title: 'Results',
      generating: 'Generating images...',
      generatingHint: 'Generation time depends on the upstream model and image count.',
      empty: 'Enter a prompt and click Generate to see your results here.',
      imageAlt: 'Generated image {index}',
      previewTitle: 'Image preview',
      previous: 'Previous image',
      next: 'Next image'
    },
    edit: {
      source: 'Reference images',
      chooseFiles: 'Choose images',
      sourceAlt: 'Reference image {index}',
      removeSource: 'Remove reference image',
      sourceRequired: 'Choose at least one reference image.',
      invalidSourceType: 'Only PNG, JPEG, or WebP images are supported.',
      sourceTooLarge: 'Each reference image must be 20 MB or smaller.',
      tooManySources: 'Choose at most {max} reference images.',
      sourceLoadFailed: 'Unable to read this generated image. Download it and upload it again.'
    },
    empty: {
      title: 'No image groups available',
      description: 'Bind an active OpenAI channel to an available group and configure a gpt-image-* model in that channel, then refresh this page.'
    },
    validation: {
      promptRequired: 'Enter a prompt.',
      promptTooLong: 'The prompt is too long. Shorten it and try again.',
      customSize: {
        positive_integer: 'Width and height must be positive integers.',
        edge_multiple: 'Width and height must satisfy the edge multiple constraint.',
        max_edge: 'Width and height exceed the model maximum edge.',
        pixel_range: 'Total pixels are outside the model-supported range.',
        aspect_ratio: 'The aspect ratio exceeds the model-supported range.'
      }
    },
    errors: {
      options: 'Unable to load image-generation options. Try again later.',
      noModel: 'Select an available group and model.',
      emptyResponse: 'The server returned no displayable image.',
      generate: 'Image generation failed. Try again later.',
      edit: 'Image editing failed. Try again later.',
      optimize: 'Prompt optimization failed. Try again later.',
      optimizeEmpty: 'The optimizer returned an empty prompt.'
    },
    settings: {
      title: 'Image Studio Settings',
      description: 'Choose the models and server-side API keys used by prompt optimization and image generation.',
      back: 'Back to Image Studio',
      promptTitle: 'Prompt optimization',
      promptHint: 'The optimizer uses the selected OpenAI-compatible model and key on the server.',
      imageTitle: 'Image generation',
      imageHint: 'The selected image model and key are used as the defaults in Image Studio.',
      group: 'Group',
      model: 'Model',
      apiKey: 'Internal API key',
      defaultSize: 'Default size',
      defaultCount: 'Default quantity',
      defaultCountHint: 'Choose between 1 and 9 images. Multiple images are shown in a grid.',
      noPromptKey: 'No active key is available for this group.',
      noImageKey: 'No active image-generation key is available for this group.',
      save: 'Save settings',
      saving: 'Saving...',
      saved: 'Settings saved.',
      loadError: 'Unable to load image-generation settings.',
      saveError: 'Unable to save image-generation settings.'
    }
  }
}
