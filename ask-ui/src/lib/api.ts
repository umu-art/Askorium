import { Configuration, SearchApi, FeedbackApi, SourceApi } from 'ts-ask-ui-api'

const config = new Configuration({
  basePath: import.meta.env.VITE_API_BASE_URL || '',
  credentials: 'include',
})

export const searchApi = new SearchApi(config)
export const feedbackApi = new FeedbackApi(config)
export const sourceApi = new SourceApi(config)

export type {
  SearchCreateRequest,
  SearchCreateResponse,
  SearchGetResponse,
  SourceSnippet,
  SourceDto,
  SourceAutoSyncPolicy,
  SourceSyncRequest,
  FeedbackDto,
} from 'ts-ask-ui-api'

export { SearchMode, SearchStatus } from 'ts-ask-ui-api'
