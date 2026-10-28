export interface VocabularyAttribution {
  title: string;
  sourceUrl: string;
  licenseUrl: string;
  wordnetLicense: string;
}

/** Approved asset for the current canonical meaning; omitted when unavailable. */
export interface MeaningImage {
  meaningId: string;
  meaningRevision: string;
  imageUrl: string;
  alt: string;
  width: number;
  height: number;
  sha256: string;
}

export interface MeaningImagesResponse {
  items: MeaningImage[];
}

export interface KnowledgeMapItem {
  meaningRevision?: string;
  image?: MeaningImage;
  rank: number;
  meaningId: string;
  wordId: string;
  wordText: string;
  wordSlug: string;
  partOfSpeech: string;
  shortDefinition: string;
  exampleSentence?: string;
  selfReportedKnown: boolean;
  saved: boolean;
  due: boolean;
  revision: string;
}

export interface KnowledgeMap {
  attribution: VocabularyAttribution;
  catalogVersion: string;
  totalWords: number;
  bandSize: number;
  ordering: { title: string; description: string; sourceUrl: string };
  bands: {
    number: number;
    startRank: number;
    endRank: number;
    totalWords: number;
    knownCount: number;
    learningCount: number;
    unseenCount: number;
  }[];
  currentBand: { number: number; items: KnowledgeMapItem[] };
}

export interface KnowledgeMapDecision {
  meaningId: string;
  decision: "learn" | "known";
  expectedRevision: string;
}

/**
 * The meaning's current server state after a decision. A replayed idempotency
 * key returns the state as it is now, which can differ from `decision` when a
 * later action (another tab, a removal) changed the meaning since.
 */
export interface KnowledgeMapDecisionResult {
  meaningId: string;
  decision: "learn" | "known";
  revision: string;
  selfReportedKnown: boolean;
  saved: boolean;
  due: boolean;
}

export interface DailyLesson {
  attribution: VocabularyAttribution;
  id: string;
  status: "in_progress" | "completed";
  revision: number;
  completedWords: number;
  totalWords: number;
  firstAnswersCorrect: number;
  questionsAnswered: number;
  creditedToday: number;
  currentStep: {
    id: string;
    kind: "meaning_choice";
    wordText: string;
    wordSlug: string;
    partOfSpeech: string;
    pronunciationMeaningId?: string;
    pronunciationRevision?: string;
    choices: { id: string; text: string }[];
  } | null;
  feedback: {
    stepId: string;
    correct: boolean;
    answer: string;
    explanation: string;
    meaningId: string;
    wordText: string;
    credited: boolean;
    skipped: boolean;
    reason?: string;
    correctChoiceId?: string;
    selectedChoiceId?: string;
    nextReviewAt?: string;
    reviewStep?: number;
  } | null;
  canContinue: boolean;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
}

export interface DailyLessonAction {
  stepId: string;
  expectedRevision: number;
  clientActionId: string;
  action: "answer" | "continue" | "skip";
  choiceId?: string;
  wasHintUsed?: boolean;
  nextReview?:
    "automatic" | "1d" | "3d" | "7d" | "14d" | "30d" | "60d" | "120d" | "known";
}

export interface CurrentUser {
  /** Opaque learner identity for local draft isolation; absent on older APIs. */
  id?: string;
  email?: string;
  displayName?: string;
  avatarUrl?: string;
  emailVerifiedAt?: string;
  /** Whether this account has an email/password credential. */
  hasPassword?: boolean;
  /**
   * VOC-031-T01 additive field. Always present in the response.
   * The Next.js middleware uses it to gate the core-loop routes
   * on whether the learner has completed onboarding (DOC-03 §3).
   */
  onboardingStatus: "not_started" | "in_progress" | "completed";
}

export type EnglishLevel = "a1" | "a2" | "b1" | "b2" | "unknown";

export type LearningGoal =
  "general" | "work" | "travel" | "study" | "conversation" | "exam";

export type MainUseCase = "daily_life" | "work" | "travel" | "study" | "social";

export interface LearningPreferences {
  learningGoal: LearningGoal | null;
  mainUseCase: MainUseCase | null;
  revision: number;
}

export interface LearningPreferencesUpdate {
  learningGoal: LearningGoal;
  mainUseCase: MainUseCase;
  expectedRevision: number;
}

export interface OnboardingProfile {
  status: "not_started" | "in_progress" | "completed";
  englishLevel?: EnglishLevel;
  nativeLanguage?: string;
  learningGoal?: LearningGoal;
  mainUseCase?: MainUseCase;
  dailyReviewTarget?: number;
  completedAt?: string;
}

export interface CompleteOnboardingBody {
  englishLevel: EnglishLevel;
  nativeLanguage: string;
  learningGoal: LearningGoal;
  mainUseCase: MainUseCase;
  dailyReviewTarget: number;
  /** Optional browser-resolved IANA timezone for learner-local daily logic. */
  timezone?: string;
}

export interface RequestMagicLinkBody {
  email: string;
  returnTo?: string;
}

export interface ConsumeMagicLinkBody {
  token: string;
  email: string;
}

export interface PasswordSignupBody {
  email: string;
  password: string;
  displayName?: string;
}

export interface VerifyPasswordSignupBody {
  token: string;
}

export interface PasswordLoginBody {
  email: string;
  password: string;
}

export interface PasswordResetRequestBody {
  email: string;
}

export interface PasswordResetBody {
  token: string;
  password: string;
}

export interface OAuthStartBody {
  redirectUri: string;
  rememberMe?: boolean;
}

export interface OAuthStartResponse {
  url: string;
}

export interface PronunciationAudio {
  audioBase64: string;
  mimeType: "audio/mpeg";
  voice: string;
  cacheHit: boolean;
  meaningId?: string;
  meaningRevision?: string;
  partOfSpeech?: string;
}

/**
 * VOC-084-T01. Kill-switch state reported by GET /healthz. Field
 * names match the API's snake_case JSON (apps/api production
 * HealthzOutput).
 */
export interface KillSwitchStatus {
  magic_link_enabled?: boolean;
  oauth_enabled?: boolean;
  new_signups_enabled?: boolean;
  ai_enabled?: boolean;
  password_enabled?: boolean;
}

/**
 * VOC-084-T01. Unauthenticated liveness probe body. The handler
 * returns HTTP 200 when healthy and 503 when unhealthy, but
 * kill_switches are present in both cases.
 */
export interface HealthzResponse {
  status: string;
  database?: string;
  timestamp?: string;
  kill_switches?: KillSwitchStatus;
}

export interface Situation {
  id: string;
  slug: string;
  title: string;
  shortDescription: string;
  levelBand?: string;
  category: string;
  displayOrder: number;
}

export interface SituationMeaning {
  meaningId: string;
  wordId: string;
  wordSlug: string;
  wordText: string;
  partOfSpeech: string;
  shortDefinition: string;
  saved: boolean;
  selfReportedKnown: boolean;
}

export interface WordExample {
  id: string;
  exampleText: string;
  situationLabel?: string;
}

export interface WordUsageNote {
  id: string;
  noteType: string;
  noteText: string;
}

export interface WordMeaning {
  meaningRevision?: string;
  image?: MeaningImage;
  id: string;
  partOfSpeech: string;
  shortDefinition: string;
  learnerDefinition?: string;
  saved: boolean;
  selfReportedKnown: boolean;
  userWordId?: string;
  reviewState?:
    "new" | "learning" | "reviewing" | "mastered" | "ignored" | "archived";
  due?: boolean;
  examples: WordExample[];
  usageNotes: WordUsageNote[];
}

export interface WordDetail {
  id: string;
  text: string;
  slug: string;
  wordType: string;
  difficultyLevel?: string;
  meanings: WordMeaning[];
}

export interface ListSituationsResponse {
  items: Situation[];
  nextCursor?: string;
  hasMore: boolean;
}

export interface SituationResponse {
  situation: Situation;
  meanings: SituationMeaning[];
}

export interface WordDetailResponse {
  word: WordDetail;
}

export interface SavedMeaning {
  image?: MeaningImage;
  userWordId: string;
  meaningId: string;
  wordId: string;
  wordText: string;
  wordSlug: string;
  partOfSpeech: string;
  shortDefinition: string;
  status: string;
  source: string;
  saved: boolean;
  addedAt: string;
}

export interface ListSavedWordsResponse {
  items: Array<SavedMeaning & { reviewState: string; due: boolean }>;
  totalCount: number;
  nextCursor?: string;
  hasMore: boolean;
}

export interface ListSavedWordsParams {
  q?: string;
  stage?:
    "" | "new" | "learning" | "reviewing" | "mastered" | "ignored" | "archived";
  due?: boolean;
  after?: string;
  limit?: number;
}

export interface VocabularySearchItem {
  meaningId: string;
  wordId: string;
  wordSlug: string;
  wordText: string;
  partOfSpeech: string;
  shortDefinition: string;
  difficultyLevel: string;
  saved: boolean;
  selfReportedKnown: boolean;
  userWordId?: string;
  reviewState?: string;
  due: boolean;
}

export interface VocabularySearchParams {
  q?: string;
  category?: string;
  level?: string;
  knowledge?: "known" | "saved" | "unexplored" | "";
  after?: string;
  limit?: number;
}

export interface VocabularySearchResponse {
  items: VocabularySearchItem[];
  totalCount: number;
  nextCursor?: string;
  hasMore: boolean;
}

/** Read-only dictionary text; these entries have no canonical learning IDs. */
export interface DictionaryEntry {
  word: string;
  meanings: Array<{
    partOfSpeech: string;
    definitions: Array<{ definition: string; example?: string }>;
  }>;
  attribution: {
    provider: string;
    providerUrl: string;
    sourceUrls: string[];
    licenses: Array<{ name: string; url: string; text: string }>;
  };
}

export interface KnowledgeSummary {
  selfReportedKnown: number;
  saved: number;
  new: number;
  learning: number;
  reviewing: number;
  mastered: number;
  ignored: number;
  archived: number;
  due: number;
}

export interface MeaningKnowledge {
  meaningId: string;
  selfReportedKnown: boolean;
  note: string;
  updatedAt?: string;
}

export interface Achievement {
  id: string;
  label: string;
  description: string;
  category: "lessons" | "practice" | "reviews" | "writing";
  criterion: "participation" | "unaided_recall";
  current: number;
  target: number;
  earned: boolean;
  earnedAt?: string;
}

export interface AchievementList {
  catalogVersion: string;
  items: Achievement[];
}

export interface MeaningKnowledgeUpdate {
  selfReportedKnown: boolean;
  note: string;
}

export interface WordListSummary {
  id: string;
  name: string;
  revision: number;
  memberCount: number;
  usableMemberCount: number;
  createdAt: string;
  updatedAt: string;
}
export interface WordListMember {
  meaningId: string;
  wordId: string;
  wordSlug: string;
  wordText: string;
  shortDefinition: string;
  partOfSpeech: string;
  practiceAvailable: boolean;
  addedAt: string;
}
export interface WordListDetail extends WordListSummary {
  members: WordListMember[];
}
export interface WordListsResponse {
  items: WordListSummary[];
}
export interface WordListWriteBody {
  name: string;
  expectedRevision: number;
}
export interface WordListRevisionBody {
  expectedRevision: number;
}

export interface StoryLine {
  id: string;
  speaker: string;
  text: string;
}
export interface StoryVocabulary {
  wordText: string;
  wordSlug: string;
  meaningId: string;
  definition: string;
}
export interface StoryChoice {
  id: string;
  text: string;
}
export interface StoryStep {
  id: string;
  kind: "line" | "comprehension" | "phrase_completion";
  line?: StoryLine;
  prompt?: string;
  choices?: StoryChoice[];
}
export interface StoryFeedback {
  stepId: string;
  correct: boolean;
  answer: string;
  explanation: string;
}
export interface StorySummary {
  id: string;
  storyKey: string;
  title: string;
  situation: string;
  contentVersion: string;
  gradingVersion: string;
  status: "in_progress" | "completed";
  revision: number;
  completedSteps: number;
  totalSteps: number;
  firstAnswersCorrect: number;
  questionsAnswered: number;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
}
export interface StorySession extends StorySummary {
  currentStep: StoryStep | null;
  feedback: StoryFeedback | null;
  canContinue: boolean;
  visibleLines: StoryLine[];
  vocabulary: StoryVocabulary[];
}
export interface StoryCatalogItem {
  key: string;
  title: string;
  description: string;
  situation: string;
  level: string;
  contentVersion: string;
  lineCount: number;
  questionCount: number;
  vocabulary: StoryVocabulary[];
  latestSession?: StorySummary;
}
export interface StoryLibrary {
  items: StoryCatalogItem[];
}
export interface StoryReading extends StoryCatalogItem {
  lines: StoryLine[];
}
export interface StoryStartRequest {
  storyKey: string;
}
export interface StoryAction {
  stepId: string;
  expectedRevision: number;
  clientActionId: string;
  action: "answer" | "continue";
  choiceId?: string;
}

export type PracticeMode =
  "typed_recall" | "listening_choice" | "mistakes" | "journey";

export interface PracticeStartRequest {
  listId?: string;
  listRevision?: number;
  mode: PracticeMode;
  lessonKey?: string;
}

export interface PracticeSessionSummary {
  journeyLesson?: number;
  newWords?: number;
  listId?: string;
  listName?: string;
  listRevision?: number;
  id: string;
  mode: PracticeMode;
  lessonKey?: string;
  contentVersion: string;
  gradingVersion: string;
  status: "in_progress" | "completed";
  revision: number;
  completedSteps: number;
  totalSteps: number;
  firstAnswersCorrect: number;
  questionsAnswered: number;
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
}

export interface PracticeStep {
  image?: MeaningImage;
  pronunciationMeaningId?: string;
  pronunciationRevision?: string;
  partOfSpeech?: string;
  id: string;
  /** Teaching-only visual context; absent from graded questions. */
  meaningId?: string;
  kind:
    | "typed_recall"
    | "listening_choice"
    | "teach"
    | "meaning_choice"
    | "gap_choice"
    | "sentence_build";
  prompt: string;
  choices: { id: string; text: string }[];
  speechText?: string;
  speechLanguage?: string;
  wordText?: string;
  context?: string;
}

export interface PracticeFeedback {
  partOfSpeech?: string;
  revealed?: boolean;
  correctTokenIds?: string[];
  selectedTokenIds?: string[];
  stepId: string;
  correct: boolean;
  /** Confirmed session choice, present only after a correct listening answer. */
  correctChoiceId?: string;
  assisted: boolean;
  answer: string;
  explanation: string;
  wordText: string;
  wordSlug: string;
  meaningId: string;
}

export interface PracticeSession extends PracticeSessionSummary {
  currentStep: PracticeStep | null;
  feedback: PracticeFeedback | null;
  canContinue: boolean;
}

export interface PracticeSessionsResponse {
  items: PracticeSessionSummary[];
  availableMistakes: number;
  journey?: {
    completedLessons: number;
    currentSession: PracticeSessionSummary | null;
    nextLesson: number;
    supportedKnownWords: number;
    totalKnownWords: number;
    todayCompletedLessons: number;
    dailyGoal: number;
  };
}

export interface PracticeAction {
  stepId: string;
  expectedRevision: number;
  clientActionId: string;
  action: "answer" | "continue" | "reveal";
  typedAnswer?: string;
  choiceId?: string;
  tokenIds?: string[];
}

export interface LessonSummary {
  key: string;
  version: string;
  title: string;
  situationSlug: string;
  situationTitle: string;
  description: string;
  wordCount: number;
  stepCount: number;
  status: "not_started" | "in_progress" | "completed";
  sessionId?: string;
  completedSteps: number;
}

export interface LessonRecommendationResponse {
  status:
    | "recommended"
    | "no_unfinished_lessons"
    | "no_useful_targets"
    | "content_unavailable";
  recommendation: null | {
    lesson: LessonSummary;
    reason: "resume" | "focus_and_useful_words" | "useful_words";
    usefulTargetCount: number;
    totalTargetCount: number;
    matchesFocus: boolean;
  };
}

export interface LessonWord {
  image?: MeaningImage;
  meaningId: string;
  wordText: string;
  wordSlug: string;
  partOfSpeech: string;
  definition: string;
  example: string;
  usageNote: string;
}

export interface LessonStep {
  pronunciationMeaningId?: string;
  pronunciationRevision?: string;
  partOfSpeech?: string;
  id: string;
  kind: "teach" | "recall" | "context" | "typed_recall" | "listening_choice";
  speechText?: string;
  speechLanguage?: string;
  word: LessonWord;
  prompt: string;
  context?: string;
  choices: { id: string; text: string }[];
}

export interface LessonSession {
  exerciseVersion?: string;
  id: string;
  lessonKey: string;
  lessonVersion: string;
  title: string;
  situationSlug: string;
  status: "in_progress" | "completed";
  revision: number;
  completedSteps: number;
  totalSteps: number;
  words: LessonWord[];
  currentStep: LessonStep | null;
  feedback: {
    stepId: string;
    correct: boolean;
    explanation: string;
    correctChoiceId: string;
    answer?: string;
  } | null;
  canContinue: boolean;
  firstAnswersCorrect: number;
  questionsAnswered: number;
  completedAt?: string;
}

export interface LessonAction {
  typedAnswer?: string;
  stepId: string;
  expectedRevision: number;
  clientActionId: string;
  action: "answer" | "continue";
  choiceId?: string;
}

export interface SaveUserWordBody {
  meaningId: string;
  source: "journey" | "search" | "manual";
}

export interface DueWord {
  userWordId: string;
  meaningId: string;
  wordId: string;
  wordText: string;
  wordSlug: string;
  partOfSpeech: string;
  shortDefinition: string;
  status: string;
  reviewStep: number;
}

export interface ListDueWordsResponse {
  items: DueWord[];
  nextCursor?: string;
  hasMore: boolean;
  totalCount: number;
}

export interface ReviewAttempt {
  attemptId: string;
  userWordId: string;
  meaningId: string;
  attemptType: string;
  promptType: "multiple_choice" | "self_check";
  result: "correct" | "incorrect" | "skipped";
  rating: "again" | "hard" | "good" | "easy" | null;
  reviewStepBefore: number;
  reviewStepAfter: number;
  answeredAt: string;
  responseTimeMs: number;
  selectedOptionMeaningId?: string;
  typedAnswer?: string;
  wasHintUsed: boolean;
  source: string;
  clientAttemptId: string;
  nextReviewAt: string;
}

export interface SubmitReviewBody {
  userWordId: string;
  meaningId: string;
  attemptType?: "review";
  promptType: "multiple_choice" | "self_check";
  result: "correct" | "incorrect" | "skipped";
  rating?: "again" | "hard" | "good" | "easy";
  answeredAt: string;
  responseTimeMs?: number;
  selectedOptionMeaningId?: string;
  typedAnswer?: string;
  wasHintUsed?: boolean;
  source?: "review" | "review_session";
  clientAttemptId: string;
  metadata?: Record<string, unknown>;
}

export interface SentenceFeedbackResult {
  feedbackId?: string;
  sentenceId?: string;
  attemptId?: string;
  targetWordId?: string;
  processingStatus?: "pending" | "completed" | "failed" | "skipped";
  status?: "correct" | "needs_improvement" | "incorrect";
  originalSentence: string;
  correctedSentence: string | null;
  /** Encouraging, honest feedback headline (maximum 60 characters). */
  headline?: string;
  explanation?: string;
  improvementTip: string | null;
  targetWordUsedCorrectly: boolean;
  grammarAcceptable: boolean;
  meaningClear: boolean;
  naturalness?: "natural" | "understandable" | "unnatural";
  missionCompleted: boolean;
  canRetry: boolean;
  reported: boolean;
  errorCode?: string;
  errorMessage?: string;
  crisisResourceMessage?: string;
  createdAt?: string;
}

export interface LearnerSentence {
  id: string;
  feedbackId?: string;
  targetWordId?: string;
  processingStatus: "pending" | "completed" | "failed" | "skipped";
  status?: "correct" | "needs_improvement" | "incorrect";
  originalSentence: string;
  correctedSentence: string | null;
  headline?: string;
  explanation?: string;
  improvementTip: string | null;
  targetWordUsedCorrectly: boolean;
  grammarAcceptable: boolean;
  meaningClear: boolean;
  naturalness?: "natural" | "understandable" | "unnatural";
  reported: boolean;
  createdAt: string;
}

export interface ListLearnerSentencesResponse {
  items: LearnerSentence[];
  nextCursor?: string;
  hasMore: boolean;
}

export interface SubmitSentenceFeedbackBody {
  sentenceText: string;
  source: "word_detail" | "review" | "daily_mission" | "free_practice";
  attemptId: string;
}

export interface ReportSentenceFeedbackBody {
  reason:
    | "already_correct"
    | "correction_changed_meaning"
    | "explanation_unclear"
    | "inappropriate"
    | "something_else";
}

export interface Streak {
  currentStreakCount: number;
  longestStreakCount: number;
  status: "active" | "at_risk" | "broken";
  graceDayBalance: number;
}

export interface DailyMission {
  localDate: string;
  timezone: string;
  reviewTarget: number;
  reviewsCompleted: number;
  newWordTarget?: number;
  newWordsCompleted?: number;
  sentencePracticeTarget?: number;
  sentencePracticesCompleted?: number;
  policyVersion: string;
  status: "open" | "completed" | "missed" | "protected";
  completedAt?: string;
  graceApplied: boolean;
  streak: Streak;
}

export interface CompletionDay {
  localDate: string;
  /** Legacy streak-credit flag: true for completed and protected days. */
  completed: boolean;
  /** Optional while older API deployments still return only completed. */
  status?: DailyMission["status"];
}

export interface Progress {
  confidencePointsBalance: number;
  streak: Streak;
  completionHistory: CompletionDay[];
}

export type ReviewIntervalPreset =
  "vocanova_default" | "wordup_like" | "custom";

/**
 * VOC-031-T02 additive type. The persisted app language
 * preference; only "en" is accepted at launch (VOC-031-D06),
 * because no i18n infrastructure exists in this repository
 * today.
 */
export type AppLanguage = "en";

/**
 * VOC-031-T02. The public Settings projection returned by
 * GET /api/v1/settings. The /api/v1/settings/account frontend
 * reads this for every editable Settings field.
 */
export interface Settings {
  dailyReviewTarget: number;
  reviewIntervalPreset: ReviewIntervalPreset;
  appLanguage: AppLanguage;
  notificationsEnabled: boolean;
  marketingEmailsEnabled: boolean;
  displayName: string;
}

/**
 * VOC-031-T02. The partial-update payload for
 * PATCH /api/v1/settings. Every field is optional; the API
 * only writes the fields the caller supplies. The DOC-07 §3
 * "no-op PATCH is a well-formed read" rule is honored, so
 * an empty body returns the current state.
 */
export interface UpdateSettingsBody {
  dailyReviewTarget?: number;
  reviewIntervalPreset?: ReviewIntervalPreset;
  appLanguage?: AppLanguage;
  notificationsEnabled?: boolean;
  marketingEmailsEnabled?: boolean;
  displayName?: string;
}

/**
 * VOC-031-T03. The request body for
 * POST /api/v1/settings/email-change-links. The new email is
 * the destination the requester wants to switch to; the
 * current sign-in address is taken from the session and is
 * never trusted from the body. The request is unconditionally
 * generic on the server side, so the registration status of
 * the new address is never observable through the request
 * outcome (anti-enumeration posture, VOC-031-D05).
 */
export interface RequestEmailChangeLinkBody {
  newEmail: string;
}

/**
 * VOC-031-T03. The request body for
 * POST /api/v1/settings/email-change-links/consume. The token
 * is the only form the requester supplies; the API never sees
 * the email itself at consume time (the server resolved it
 * when the request was issued).
 */
export interface ConsumeEmailChangeLinkBody {
  token: string;
}

/**
 * VOC-031-T03. The post-confirm response from
 * POST /api/v1/settings/email-change-links/consume. The
 * server returns the new email and the previous email so the
 * frontend can show the learner which address the security
 * notification was dispatched to; the notification itself is
 * owned by the backend, not the frontend.
 */
export interface ConsumeEmailChangeLinkResult {
  email: string;
  previousEmail: string;
  changedAt: string;
}

/**
 * VOC-031-T04. The post-deactivation response from
 * POST /api/v1/account-deletion-requests. The user is already
 * deactivated at this point: status is 'deactivated', every
 * active session and every unconsumed auth/email-change
 * token is revoked, and the purge_after clock is running. The
 * frontend uses the dates to render a clear "your account
 * has been scheduled for deletion" confirmation and to
 * initiate logout. `replayed` is true when the call was a
 * no-op because the (user, idempotency-key) pair already
 * matched a prior request — the frontend uses it to
 * suppress duplicate toasts on a retry.
 */
export interface CreateAccountDeletionRequestResult {
  status: string;
  userId: string;
  requestedAt: string;
  purgeAfter: string;
  idempotencyKey: string;
  replayed: boolean;
}

/** Learner-visible history, without private grading snapshots or receipts. */
export interface PersonalDataSessionProgress {
  id: string;
  status: string;
  completedSteps: number;
  totalSteps: number;
  firstAnswersCorrect: number;
  questionsAnswered: number;
  startedAt: string;
  updatedAt: string;
  completedAt: string | null;
}
export interface PersonalDataLearningAction {
  stepId?: string;
  action?: string;
  choiceId?: string;
  typedAnswer?: string;
  lessonKey?: string;
  tokenIds?: string[];
}
export interface PersonalDataWordList {
  id: string;
  name: string;
  revision: number;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
  members: Array<{ meaningId: string; addedAt: string }>;
}
export interface PersonalDataStorySession extends PersonalDataSessionProgress {
  storyKey: string;
  title: string;
  contentVersion: string;
  gradingVersion: string;
  actions: Array<{
    action: PersonalDataLearningAction;
    feedback: StoryFeedback | null;
    createdAt: string;
  }>;
}
export interface PersonalDataPracticeSession extends PersonalDataSessionProgress {
  mode: string;
  lessonKey: string | null;
  listId: string | null;
  listName: string | null;
  listRevision: number | null;
  journeyLesson: number | null;
  newWords: number | null;
  contentVersion: string;
  gradingVersion: string;
  actions: Array<{
    action: PersonalDataLearningAction;
    meaningId: string | null;
    correct: boolean | null;
    feedback: PracticeFeedback | null;
    createdAt: string;
  }>;
}
export interface PersonalDataGuidedLesson extends PersonalDataSessionProgress {
  lessonKey: string;
  lessonVersion: string;
  exerciseVersion: string | null;
  title: string;
  words: LessonWord[];
  feedback: LessonSession["feedback"];
  actions: Array<{
    action: PersonalDataLearningAction;
    feedback: LessonSession["feedback"];
    completedSteps: number;
    createdAt: string;
  }>;
}

/** A portable learner-visible personal-data export. Values are deliberately
 * data-shaped rather than internal API DTOs so additions remain backward
 * compatible. It never contains credentials, hidden prompts, or abuse data. */
export interface PersonalDataExport {
  schemaVersion: string;
  exportedAt?: string;
  profile: Record<string, unknown>;
  settings: Record<string, unknown>;
  onboardingProfile: Record<string, unknown> | null;
  learningPreferences?:
    (LearningPreferences & { createdAt: string; updatedAt: string }) | null;
  savedWords: unknown[];
  reviewHistory: unknown[];
  sentenceFeedbackHistory: unknown[];
  dailyMissions: unknown[];
  dailyActivity: unknown[];
  confidencePointLedger: unknown[];
  graceDayLedger: unknown[];
  streakState: Record<string, unknown> | null;
  guidedLessons?: PersonalDataGuidedLesson[] | null;
  dailyLessons?: Array<{
    id: string;
    status: DailyLesson["status"];
    revision: number;
    completedWords: number;
    totalWords: number;
    firstAnswersCorrect: number;
    questionsAnswered: number;
    creditedToday: number;
    startedAt: string;
    updatedAt: string;
    completedAt?: string | null;
    actions: Array<{
      action: Pick<
        DailyLessonAction,
        "stepId" | "action" | "choiceId" | "wasHintUsed" | "nextReview"
      >;
      feedback: DailyLesson["feedback"];
      createdAt: string;
    }>;
  }> | null;
  dailyReviewSuccesses?: Array<{
    meaningId: string;
    localDate: string;
    reviewAttemptId: string;
    dailySessionId: string | null;
    createdAt: string;
  }> | null;
  wordKnowledge?: Array<{
    meaningId: string;
    selfReportedKnown: boolean;
    note: string;
    updatedAt: string;
  }> | null;
  practiceSessions?: PersonalDataPracticeSession[] | null;
  wordLists?: PersonalDataWordList[] | null;
  storySessions?: PersonalDataStorySession[] | null;
}

export interface ApiError {
  type?: string;
  title?: string;
  status?: number;
  detail?: string;
  instance?: string;
  errors?: Array<{ location?: string; message?: string; value?: unknown }>;
}

export class ApiResponseError extends Error {
  constructor(
    readonly status: number,
    readonly body: ApiError | null,
    message?: string,
  ) {
    super(message ?? body?.detail ?? `HTTP ${status}`);
  }
}

export interface VocanovaClientOptions {
  baseURL: string;
  credentials?: RequestCredentials;
  fetch?: typeof fetch;
}

export class VocanovaClient {
  private readonly fetch: typeof fetch;

  constructor(private readonly options: VocanovaClientOptions) {
    this.fetch = options.fetch ?? globalThis.fetch.bind(globalThis);
  }

  /**
   * VOC-084-T01. Fetch GET /healthz for deploy-derived capability
   * signals (kill switches). Unlike other client methods, this does
   * not throw on HTTP 503 so callers can still read kill_switches
   * when the database probe is unhealthy.
   */
  async getHealthz(init?: RequestInit): Promise<{
    data: HealthzResponse;
    response: Response;
  }> {
    const url = new URL("/healthz", this.options.baseURL);
    const headers = new Headers(init?.headers);
    if (!headers.has("Accept")) {
      headers.set("Accept", "application/json");
    }

    const response = await this.fetch(url.toString(), {
      ...init,
      method: "GET",
      headers,
      credentials: this.options.credentials,
    });
    const data = (await response.json()) as HealthzResponse;
    return { data, response };
  }

  async getCurrentUser(init?: RequestInit): Promise<{
    data: CurrentUser;
    response: Response;
  }> {
    const response = await this.request("GET", "/api/v1/me", undefined, init);
    const data = (await response.json()) as CurrentUser;
    return { data, response };
  }

  async getOnboarding(init?: RequestInit): Promise<{
    data: OnboardingProfile;
    response: Response;
  }> {
    const response = await this.request(
      "GET",
      "/api/v1/onboarding",
      undefined,
      init,
    );
    const data = (await response.json()) as OnboardingProfile;
    return { data, response };
  }

  async completeOnboarding(
    body: CompleteOnboardingBody,
    init?: RequestInit,
  ): Promise<{ data: OnboardingProfile; response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/onboarding",
      body,
      init,
    );
    const data = (await response.json()) as OnboardingProfile;
    return { data, response };
  }

  async requestMagicLink(
    body: RequestMagicLinkBody,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/magic-links",
      body,
      init,
    );
    return { response };
  }

  async consumeMagicLink(
    body: ConsumeMagicLinkBody,
    init?: RequestInit,
  ): Promise<{ data: CurrentUser; response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/magic-links/consume",
      body,
      init,
    );
    const data = (await response.json()) as CurrentUser;
    return { data, response };
  }

  async requestPasswordSignup(
    body: PasswordSignupBody,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/password/signups",
      body,
      init,
    );
    return { response };
  }

  async verifyPasswordSignup(
    body: VerifyPasswordSignupBody,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/password/signups/verify",
      body,
      init,
    );
    return { response };
  }

  async loginWithPassword(
    body: PasswordLoginBody,
    init?: RequestInit,
  ): Promise<{ data: CurrentUser; response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/password/login",
      body,
      init,
    );
    const data = (await response.json()) as CurrentUser;
    return { data, response };
  }

  async requestPasswordReset(
    body: PasswordResetRequestBody,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/password/reset-requests",
      body,
      init,
    );
    return { response };
  }

  async resetPassword(
    body: PasswordResetBody,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/password/resets",
      body,
      init,
    );
    return { response };
  }

  async startOAuth(
    body: OAuthStartBody,
    init?: RequestInit,
  ): Promise<{ data: OAuthStartResponse; response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/oauth/google/start",
      body,
      init,
    );
    const data = (await response.json()) as OAuthStartResponse;
    return { data, response };
  }

  async logout(init?: RequestInit): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/auth/logout",
      undefined,
      init,
    );
    return { response };
  }

  async getPronunciation(
    text: string,
    init?: RequestInit,
    meaningId?: string,
    meaningRevision?: string,
  ): Promise<{ data: PronunciationAudio; response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/pronunciation",
      {
        text,
        ...(meaningId ? { meaningId } : {}),
        ...(meaningRevision ? { meaningRevision } : {}),
      },
      init,
    );
    return { data: (await response.json()) as PronunciationAudio, response };
  }

  async listJourneySituations(
    params?: { after?: string; limit?: number },
    init?: RequestInit,
  ): Promise<{ data: ListSituationsResponse; response: Response }> {
    const query = new URLSearchParams();
    if (params?.after) {
      query.set("after", params.after);
    }
    if (params?.limit !== undefined) {
      query.set("limit", String(params.limit));
    }
    const path =
      "/api/v1/journey-situations" +
      (query.toString() ? `?${query.toString()}` : "");
    const response = await this.request("GET", path, undefined, init);
    const data = (await response.json()) as ListSituationsResponse;
    return { data, response };
  }

  async lookupDictionary(
    word: string,
    init?: RequestInit,
  ): Promise<{ data: DictionaryEntry; response: Response }> {
    const query = new URLSearchParams({ q: word });
    const response = await this.request(
      "GET",
      `/api/v1/dictionary?${query.toString()}`,
      undefined,
      init,
    );
    return { data: (await response.json()) as DictionaryEntry, response };
  }

  async searchVocabulary(
    params: VocabularySearchParams = {},
    init?: RequestInit,
  ): Promise<{ data: VocabularySearchResponse; response: Response }> {
    const query = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined && value !== "") query.set(key, String(value));
    }
    const response = await this.request(
      "GET",
      `/api/v1/canonical-words${query.size ? `?${query}` : ""}`,
      undefined,
      init,
    );
    return {
      data: (await response.json()) as VocabularySearchResponse,
      response,
    };
  }

  async getKnowledgeSummary(
    init?: RequestInit,
  ): Promise<{ data: KnowledgeSummary; response: Response }> {
    const response = await this.request(
      "GET",
      "/api/v1/knowledge-summary",
      undefined,
      init,
    );
    return { data: (await response.json()) as KnowledgeSummary, response };
  }

  async listWordLists(
    init?: RequestInit,
  ): Promise<{ data: WordListsResponse; response: Response }> {
    const response = await this.request(
      "GET",
      "/api/v1/word-lists",
      undefined,
      init,
    );
    return { data: (await response.json()) as WordListsResponse, response };
  }

  async getWordList(
    listId: string,
    init?: RequestInit,
  ): Promise<{ data: WordListDetail; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/word-lists/${encodeURIComponent(listId)}`,
      undefined,
      init,
    );
    return { data: (await response.json()) as WordListDetail, response };
  }

  async putWordList(
    listId: string,
    body: WordListWriteBody,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: WordListDetail; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "PUT",
      `/api/v1/word-lists/${encodeURIComponent(listId)}`,
      body,
      { ...init, headers },
    );
    return { data: (await response.json()) as WordListDetail, response };
  }

  async deleteWordList(
    listId: string,
    expectedRevision: number,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "DELETE",
      `/api/v1/word-lists/${encodeURIComponent(listId)}?expectedRevision=${expectedRevision}`,
      undefined,
      { ...init, headers },
    );
    return { response };
  }

  async putWordListMember(
    listId: string,
    meaningId: string,
    body: WordListRevisionBody,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: WordListDetail; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "PUT",
      `/api/v1/word-lists/${encodeURIComponent(listId)}/members/${encodeURIComponent(meaningId)}`,
      body,
      { ...init, headers },
    );
    return { data: (await response.json()) as WordListDetail, response };
  }

  async deleteWordListMember(
    listId: string,
    meaningId: string,
    expectedRevision: number,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: WordListDetail; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "DELETE",
      `/api/v1/word-lists/${encodeURIComponent(listId)}/members/${encodeURIComponent(meaningId)}?expectedRevision=${expectedRevision}`,
      undefined,
      { ...init, headers },
    );
    return { data: (await response.json()) as WordListDetail, response };
  }

  async listStories(
    init?: RequestInit,
  ): Promise<{ data: StoryLibrary; response: Response }> {
    const response = await this.request(
      "GET",
      "/api/v1/stories",
      undefined,
      init,
    );
    return { data: (await response.json()) as StoryLibrary, response };
  }

  async getStory(
    storyKey: string,
    init?: RequestInit,
  ): Promise<{ data: StoryReading; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/stories/${encodeURIComponent(storyKey)}`,
      undefined,
      init,
    );
    return { data: (await response.json()) as StoryReading, response };
  }

  async startStorySession(
    body: StoryStartRequest,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: StorySession; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/story-sessions",
      body,
      { ...init, headers },
    );
    return { data: (await response.json()) as StorySession, response };
  }

  async getStorySession(
    sessionId: string,
    init?: RequestInit,
  ): Promise<{ data: StorySession; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/story-sessions/${encodeURIComponent(sessionId)}`,
      undefined,
      init,
    );
    return { data: (await response.json()) as StorySession, response };
  }

  async submitStoryAction(
    sessionId: string,
    body: StoryAction,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: StorySession; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      `/api/v1/story-sessions/${encodeURIComponent(sessionId)}/actions`,
      body,
      { ...init, headers },
    );
    return { data: (await response.json()) as StorySession, response };
  }

  async listPracticeSessions(
    init?: RequestInit,
  ): Promise<{ data: PracticeSessionsResponse; response: Response }> {
    const response = await this.request(
      "GET",
      "/api/v1/practice-sessions",
      undefined,
      init,
    );
    return {
      data: (await response.json()) as PracticeSessionsResponse,
      response,
    };
  }

  async listAchievements(
    init?: RequestInit,
  ): Promise<{ data: AchievementList; response: Response }> {
    const response = await this.request(
      "GET",
      "/api/v1/achievements",
      undefined,
      init,
    );
    return { data: (await response.json()) as AchievementList, response };
  }

  async getLearningPreferences(
    init?: RequestInit,
  ): Promise<{ data: LearningPreferences; response: Response }> {
    const response = await this.request(
      "GET",
      "/api/v1/learning-preferences",
      undefined,
      init,
    );
    return { data: (await response.json()) as LearningPreferences, response };
  }

  async updateLearningPreferences(
    body: LearningPreferencesUpdate,
    init?: RequestInit,
  ): Promise<{ data: LearningPreferences; response: Response }> {
    const response = await this.request(
      "PATCH",
      "/api/v1/learning-preferences",
      body,
      init,
    );
    return { data: (await response.json()) as LearningPreferences, response };
  }

  async startPracticeSession(
    body: PracticeStartRequest,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: PracticeSession; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/practice-sessions",
      body,
      { ...init, headers },
    );
    return { data: (await response.json()) as PracticeSession, response };
  }

  async getPracticeSession(
    sessionId: string,
    init?: RequestInit,
  ): Promise<{ data: PracticeSession; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/practice-sessions/${encodeURIComponent(sessionId)}`,
      undefined,
      init,
    );
    return { data: (await response.json()) as PracticeSession, response };
  }

  async submitPracticeAction(
    sessionId: string,
    body: PracticeAction,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: PracticeSession; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      `/api/v1/practice-sessions/${encodeURIComponent(sessionId)}/actions`,
      body,
      { ...init, headers },
    );
    return { data: (await response.json()) as PracticeSession, response };
  }

  async listLessons(
    init?: RequestInit,
  ): Promise<{ data: { items: LessonSummary[] }; response: Response }> {
    const response = await this.request(
      "GET",
      "/api/v1/lessons",
      undefined,
      init,
    );
    return {
      data: (await response.json()) as { items: LessonSummary[] },
      response,
    };
  }

  async getLessonRecommendation(
    init?: RequestInit,
  ): Promise<{ data: LessonRecommendationResponse; response: Response }> {
    const response = await this.request(
      "GET",
      "/api/v1/lesson-recommendation",
      undefined,
      init,
    );
    return {
      data: (await response.json()) as LessonRecommendationResponse,
      response,
    };
  }

  async getMeaningKnowledge(
    meaningId: string,
    init?: RequestInit,
  ): Promise<{ data: MeaningKnowledge; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/meaning-knowledge/${encodeURIComponent(meaningId)}`,
      undefined,
      init,
    );
    return { data: (await response.json()) as MeaningKnowledge, response };
  }

  async updateMeaningKnowledge(
    meaningId: string,
    body: MeaningKnowledgeUpdate,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: MeaningKnowledge; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "PUT",
      `/api/v1/meaning-knowledge/${encodeURIComponent(meaningId)}`,
      body,
      { ...init, headers },
    );
    return { data: (await response.json()) as MeaningKnowledge, response };
  }

  /** Change only the self-assessment; the server preserves the current note. */
  async setMeaningKnown(
    meaningId: string,
    selfReportedKnown: boolean,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: MeaningKnowledge; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "PATCH",
      `/api/v1/meaning-knowledge/${encodeURIComponent(meaningId)}`,
      { selfReportedKnown },
      { ...init, headers },
    );
    return { data: (await response.json()) as MeaningKnowledge, response };
  }

  async clearMeaningKnowledge(
    meaningId: string,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "DELETE",
      `/api/v1/meaning-knowledge/${encodeURIComponent(meaningId)}`,
      undefined,
      { ...init, headers },
    );
    return { response };
  }

  /** Bounded teaching-image lookup; never load the full catalog into the client. */
  async getMeaningImages(
    meaningIds: string[],
    init?: RequestInit,
  ): Promise<{ data: MeaningImagesResponse; response: Response }> {
    const query = new URLSearchParams({ meaningIds: meaningIds.join(",") });
    const response = await this.request(
      "GET",
      `/api/v1/meaning-images?${query.toString()}`,
      undefined,
      init,
    );
    return { data: await response.json(), response };
  }

  async getKnowledgeMap(
    band = 1,
    init?: RequestInit,
  ): Promise<{ data: KnowledgeMap; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/knowledge-map?band=${encodeURIComponent(String(band))}`,
      undefined,
      init,
    );
    return { data: await response.json(), response };
  }

  async decideKnowledgeMapWord(
    body: KnowledgeMapDecision,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: KnowledgeMapDecisionResult; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/knowledge-map/decisions",
      body,
      { ...init, headers },
    );
    return { data: await response.json(), response };
  }

  async getCurrentDailyLesson(init?: RequestInit): Promise<{
    data: { session: DailyLesson | null; dueCount: number };
    response: Response;
  }> {
    const response = await this.request(
      "GET",
      "/api/v1/daily-lessons/current",
      undefined,
      init,
    );
    return { data: await response.json(), response };
  }

  async startDailyLesson(
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: DailyLesson; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/daily-lessons",
      {},
      { ...init, headers },
    );
    return { data: await response.json(), response };
  }

  async getDailyLesson(
    sessionId: string,
    init?: RequestInit,
  ): Promise<{ data: DailyLesson; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/daily-lessons/${encodeURIComponent(sessionId)}`,
      undefined,
      init,
    );
    return { data: await response.json(), response };
  }

  async submitDailyLessonAction(
    sessionId: string,
    body: DailyLessonAction,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: DailyLesson; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      `/api/v1/daily-lessons/${encodeURIComponent(sessionId)}/actions`,
      body,
      { ...init, headers },
    );
    return { data: await response.json(), response };
  }

  async startLesson(
    lessonKey: string,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: LessonSession; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      `/api/v1/lessons/${encodeURIComponent(lessonKey)}/sessions`,
      undefined,
      { ...init, headers },
    );
    return { data: (await response.json()) as LessonSession, response };
  }

  async getLessonSession(
    sessionId: string,
    init?: RequestInit,
  ): Promise<{ data: LessonSession; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/lesson-sessions/${encodeURIComponent(sessionId)}`,
      undefined,
      init,
    );
    return { data: (await response.json()) as LessonSession, response };
  }

  async submitLessonAction(
    sessionId: string,
    body: LessonAction,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: LessonSession; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      `/api/v1/lesson-sessions/${encodeURIComponent(sessionId)}/actions`,
      body,
      { ...init, headers },
    );
    return { data: (await response.json()) as LessonSession, response };
  }

  async getJourneySituation(
    slug: string,
    init?: RequestInit,
  ): Promise<{ data: SituationResponse; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/journey-situations/${encodeURIComponent(slug)}`,
      undefined,
      init,
    );
    const data = (await response.json()) as SituationResponse;
    return { data, response };
  }

  async getCanonicalWord(
    wordSlug: string,
    init?: RequestInit,
  ): Promise<{ data: WordDetailResponse; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/canonical-words/${encodeURIComponent(wordSlug)}`,
      undefined,
      init,
    );
    const data = (await response.json()) as WordDetailResponse;
    return { data, response };
  }

  async listSavedWords(
    params?: ListSavedWordsParams,
    init?: RequestInit,
  ): Promise<{ data: ListSavedWordsResponse; response: Response }> {
    const query = new URLSearchParams();
    if (params?.q) query.set("q", params.q);
    if (params?.stage) query.set("stage", params.stage);
    if (params?.due) query.set("due", "true");
    if (params?.after) {
      query.set("after", params.after);
    }
    if (params?.limit !== undefined) {
      query.set("limit", String(params.limit));
    }
    const path =
      "/api/v1/user-words" + (query.toString() ? `?${query.toString()}` : "");
    const response = await this.request("GET", path, undefined, init);
    const data = (await response.json()) as ListSavedWordsResponse;
    return { data, response };
  }

  async getSavedWord(
    userWordId: string,
    init?: RequestInit,
  ): Promise<{ data: SavedMeaning; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/user-words/records/${encodeURIComponent(userWordId)}`,
      undefined,
      init,
    );
    const data = (await response.json()) as SavedMeaning;
    return { data, response };
  }

  async listDueWords(
    params?: { after?: string; limit?: number },
    init?: RequestInit,
  ): Promise<{ data: ListDueWordsResponse; response: Response }> {
    const query = new URLSearchParams();
    if (params?.after) {
      query.set("after", params.after);
    }
    if (params?.limit !== undefined) {
      query.set("limit", String(params.limit));
    }
    const path =
      "/api/v1/reviews/due" + (query.toString() ? `?${query.toString()}` : "");
    const response = await this.request("GET", path, undefined, init);
    const data = (await response.json()) as ListDueWordsResponse;
    return { data, response };
  }

  async saveUserWord(
    body: SaveUserWordBody,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: SavedMeaning; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request("POST", "/api/v1/user-words", body, {
      ...init,
      headers,
    });
    const data = (await response.json()) as SavedMeaning;
    return { data, response };
  }

  async unsaveUserWord(
    meaningId: string,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "DELETE",
      `/api/v1/user-words/${encodeURIComponent(meaningId)}`,
      undefined,
      init,
    );
    return { response };
  }

  async submitReview(
    body: SubmitReviewBody,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: ReviewAttempt; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/reviews/submissions",
      body,
      {
        ...init,
        headers,
      },
    );
    const data = (await response.json()) as ReviewAttempt;
    return { data, response };
  }

  async submitSentenceFeedback(
    body: SubmitSentenceFeedbackBody,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: SentenceFeedbackResult; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/learner-sentences",
      body,
      {
        ...init,
        headers,
      },
    );
    const data = (await response.json()) as SentenceFeedbackResult;
    return { data, response };
  }

  async listLearnerSentences(
    params?: { after?: string; limit?: number },
    init?: RequestInit,
  ): Promise<{ data: ListLearnerSentencesResponse; response: Response }> {
    const query = new URLSearchParams();
    if (params?.after) query.set("after", params.after);
    if (params?.limit !== undefined) query.set("limit", String(params.limit));
    const suffix = query.toString();
    const response = await this.request(
      "GET",
      `/api/v1/learner-sentences${suffix ? `?${suffix}` : ""}`,
      undefined,
      init,
    );
    const data = (await response.json()) as ListLearnerSentencesResponse;
    return { data, response };
  }

  async getLearnerSentence(
    sentenceId: string,
    init?: RequestInit,
  ): Promise<{ data: LearnerSentence; response: Response }> {
    const response = await this.request(
      "GET",
      `/api/v1/learner-sentences/${encodeURIComponent(sentenceId)}`,
      undefined,
      init,
    );
    const data = (await response.json()) as LearnerSentence;
    return { data, response };
  }

  async reportSentenceFeedback(
    attemptId: string,
    body: ReportSentenceFeedbackBody,
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      `/api/v1/sentence-feedback/${encodeURIComponent(attemptId)}/reports`,
      body,
      { ...init, headers },
    );
    return { response };
  }

  async getDailyMission(
    params?: { timezone?: string },
    init?: RequestInit,
  ): Promise<{ data: DailyMission; response: Response }> {
    const query = new URLSearchParams();
    if (params?.timezone) {
      query.set("timezone", params.timezone);
    }
    const path =
      "/api/v1/daily-mission" +
      (query.toString() ? `?${query.toString()}` : "");
    const response = await this.request("GET", path, undefined, init);
    const data = (await response.json()) as DailyMission;
    return { data, response };
  }

  async getProgress(
    params?: { timezone?: string },
    init?: RequestInit,
  ): Promise<{ data: Progress; response: Response }> {
    const query = new URLSearchParams();
    if (params?.timezone) {
      query.set("timezone", params.timezone);
    }
    const path =
      "/api/v1/progress" + (query.toString() ? `?${query.toString()}` : "");
    const response = await this.request("GET", path, undefined, init);
    const data = (await response.json()) as Progress;
    return { data, response };
  }

  /**
   * VOC-031-T02. Get the requester's settings. The response is
   * a stable Settings projection — every field is always
   * present, with schema defaults for any unset value.
   */
  async getSettings(init?: RequestInit): Promise<{
    data: Settings;
    response: Response;
  }> {
    const response = await this.request(
      "GET",
      "/api/v1/settings",
      undefined,
      init,
    );
    const data = (await response.json()) as Settings;
    return { data, response };
  }

  /**
   * VOC-031-T02. Update the requester's settings via a partial
   * PATCH. Only the fields supplied in `body` are written; every
   * other field is preserved. An empty body is a well-formed
   * no-op read and returns the current state. The `init.headers`
   * are forwarded so the caller can attach a CSRF token.
   */
  async updateSettings(
    body: UpdateSettingsBody,
    init?: RequestInit,
  ): Promise<{ data: Settings; response: Response }> {
    const response = await this.request(
      "PATCH",
      "/api/v1/settings",
      body,
      init,
    );
    const data = (await response.json()) as Settings;
    return { data, response };
  }

  /**
   * VOC-031-T03. Request a single-use email-change link. The
   * request is unconditionally generic on the server side
   * (anti-enumeration posture, VOC-031-D05): whether the
   * requested new email is already registered is never
   * observable through this response. The `init.headers` are
   * forwarded so the caller can attach a CSRF token.
   */
  async requestEmailChangeLink(
    body: RequestEmailChangeLinkBody,
    init?: RequestInit,
  ): Promise<{ response: Response }> {
    const response = await this.request(
      "POST",
      "/api/v1/settings/email-change-links",
      body,
      init,
    );
    return { response };
  }

  /**
   * VOC-031-T03. Consume a single-use email-change link. The
   * server validates the token's hash, expiry, single-use
   * `consumed_at`, and environment, re-checks new-email
   * uniqueness atomically at confirm time, and updates
   * `users.email`. The `init.headers` are forwarded so the
   * caller can attach a CSRF token.
   */
  async consumeEmailChangeLink(
    body: ConsumeEmailChangeLinkBody,
    init?: RequestInit,
  ): Promise<{
    data: ConsumeEmailChangeLinkResult;
    response: Response;
  }> {
    const response = await this.request(
      "POST",
      "/api/v1/settings/email-change-links/consume",
      body,
      init,
    );
    const data = (await response.json()) as ConsumeEmailChangeLinkResult;
    return { data, response };
  }

  /**
   * VOC-031-T04. Deactivate the requester's account and
   * schedule anonymization. Requires a CSRF token and a
   * unique Idempotency-Key (DOC-07). A replay with the same
   * key returns the existing row with `replayed: true`, so
   * the frontend can suppress duplicate "your account was
   * deleted" toasts on a retry. The user is already
   * deactivated at this point: every active session is
   * revoked, and the `purgeAfter` clock is running. The
   * frontend should follow up with a logout request to clear
   * the session cookie.
   */
  async createAccountDeletionRequest(
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{
    data: CreateAccountDeletionRequestResult;
    response: Response;
  }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/account-deletion-requests",
      undefined,
      { ...init, headers },
    );
    const data = (await response.json()) as CreateAccountDeletionRequestResult;
    return { data, response };
  }

  /** Downloads the current requester's personal data synchronously as JSON.
   * Requires the same CSRF and Idempotency-Key safeguards as account deletion. */
  async exportPersonalData(
    idempotencyKey: string,
    init?: RequestInit,
  ): Promise<{ data: PersonalDataExport; response: Response }> {
    const headers = new Headers(init?.headers);
    headers.set("Idempotency-Key", idempotencyKey);
    const response = await this.request(
      "POST",
      "/api/v1/personal-data-export",
      undefined,
      { ...init, headers },
    );
    const data = (await response.json()) as PersonalDataExport;
    return { data, response };
  }

  private async request(
    method: string,
    path: string,
    body?: unknown,
    init?: RequestInit,
  ): Promise<Response> {
    const url = new URL(path, this.options.baseURL);
    const headers = new Headers(init?.headers);
    if (!headers.has("Accept")) {
      headers.set("Accept", "application/json");
    }
    if (body !== undefined && !headers.has("Content-Type")) {
      headers.set("Content-Type", "application/json");
    }

    const response = await this.fetch(url.toString(), {
      ...init,
      method,
      headers,
      credentials: this.options.credentials,
      body: body === undefined ? undefined : JSON.stringify(body),
    });

    if (!response.ok) {
      const apiError = await parseProblemDetails(response).catch(() => null);
      throw new ApiResponseError(
        response.status,
        apiError,
        apiError?.detail ?? `HTTP ${response.status}`,
      );
    }

    return response;
  }
}

async function parseProblemDetails(
  response: Response,
): Promise<ApiError | null> {
  const contentType = response.headers.get("Content-Type") ?? "";
  if (!contentType.includes("application/problem+json")) {
    return null;
  }
  try {
    return (await response.json()) as ApiError;
  } catch {
    return null;
  }
}
