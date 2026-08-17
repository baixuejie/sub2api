export type HarnessInstallStatus =
  | 'awaiting_helper'
  | 'checking_environment'
  | 'installing'
  | 'configuring'
  | 'starting'
  | 'completed'
  | 'failed'
  | 'expired'

export interface HarnessModelOption {
  id: string
  name: string
  context_window: number
  max_tokens: number
}

export interface HarnessInstallProfile {
  api_key_id: number
  api_key_name: string
  key_hint: string
  group_name: string
  platform: string
  provider: string
  provider_name: string
  protocol: string
  base_url: string
  default_model: string
  selected_model: string
  available_models: HarnessModelOption[]
}

export interface HarnessHelperDownloads {
  windows_amd64: string
  windows_arm64: string
  darwin_amd64: string
  darwin_arm64: string
  linux_amd64: string
  linux_arm64: string
  releases_page: string
}

export interface HarnessProfileResponse {
  profile: HarnessInstallProfile
  helper_downloads: HarnessHelperDownloads
  required_node: string
  dsh_version: string
}

export interface HarnessInstallSession {
  id: string
  profile: HarnessInstallProfile
  status: HarnessInstallStatus
  stage: string
  message: string
  progress: number
  harness_url?: string
  error_code?: string
  launch_uri?: string
  ticket_expires_at?: string
  created_at: string
  updated_at: string
  expires_at: string
}

export interface CreateHarnessSessionRequest {
  api_key_id: number
  model?: string
}
