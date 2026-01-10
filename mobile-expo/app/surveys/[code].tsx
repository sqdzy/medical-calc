import { useState, useEffect, useRef } from 'react';
import {
  View,
  Text,
  ScrollView,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  Switch,
  TextInput,
  Alert,
  Modal,
  FlatList,
} from 'react-native';
import { useLocalSearchParams, useNavigation, useRouter } from 'expo-router';
import { useQuery, useMutation } from '@tanstack/react-query';
import Slider from '@react-native-community/slider';
import FontAwesome from '@expo/vector-icons/FontAwesome';
import { surveysApi } from '../../src/api/client';
import type { SurveyTemplate, SurveyQuestion, SurveyAnswer, SurveySection, SurveyResult, AIAdviceResult } from '../../src/types';

function flattenQuestions(sections: SurveySection[]): SurveyQuestion[] {
  const out: SurveyQuestion[] = [];
  for (const s of sections || []) {
    for (const q of s.questions || []) {
      out.push(q);
    }
  }
  return out;
}

function getMin(q: SurveyQuestion) {
  return typeof q.min === 'number' ? q.min : 0;
}

function getMax(q: SurveyQuestion) {
  return typeof q.max === 'number' ? q.max : 10;
}

interface SelectOption {
  value: number;
  label: string;
}

interface SelectInputProps {
  options: SelectOption[];
  value: number;
  onChange: (value: number) => void;
}

function SelectInput({ options, value, onChange }: SelectInputProps) {
  const [modalVisible, setModalVisible] = useState(false);
  const selectedOption = options.find(o => o.value === value);

  return (
    <>
      <TouchableOpacity 
        style={selectStyles.trigger} 
        onPress={() => setModalVisible(true)}
      >
        <Text style={selectStyles.triggerText}>
          {selectedOption?.label || 'Выберите...'}
        </Text>
        <FontAwesome name="chevron-down" size={14} color="#6b7280" />
      </TouchableOpacity>

      <Modal
        visible={modalVisible}
        transparent
        animationType="slide"
        onRequestClose={() => setModalVisible(false)}
      >
        <View style={selectStyles.overlay}>
          <View style={selectStyles.modal}>
            <View style={selectStyles.modalHeader}>
              <Text style={selectStyles.modalTitle}>Выберите значение</Text>
              <TouchableOpacity onPress={() => setModalVisible(false)}>
                <FontAwesome name="times" size={24} color="#6b7280" />
              </TouchableOpacity>
            </View>
            <FlatList
              data={options}
              keyExtractor={(item) => String(item.value)}
              renderItem={({ item }) => (
                <TouchableOpacity
                  style={[
                    selectStyles.option,
                    item.value === value && selectStyles.optionSelected
                  ]}
                  onPress={() => {
                    onChange(item.value);
                    setModalVisible(false);
                  }}
                >
                  <Text style={[
                    selectStyles.optionText,
                    item.value === value && selectStyles.optionTextSelected
                  ]}>
                    {item.label}
                  </Text>
                  {item.value === value && (
                    <FontAwesome name="check" size={16} color="#2563eb" />
                  )}
                </TouchableOpacity>
              )}
            />
          </View>
        </View>
      </Modal>
    </>
  );
}

const selectStyles = StyleSheet.create({
  trigger: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: '#f9fafb',
    borderWidth: 1,
    borderColor: '#e5e7eb',
    borderRadius: 8,
    padding: 16,
    marginTop: 12,
  },
  triggerText: {
    fontSize: 16,
    color: '#374151',
    flex: 1,
  },
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'flex-end',
  },
  modal: {
    backgroundColor: '#fff',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    maxHeight: '70%',
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#e5e7eb',
  },
  modalTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1f2937',
  },
  option: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#f3f4f6',
  },
  optionSelected: {
    backgroundColor: '#eff6ff',
  },
  optionText: {
    fontSize: 16,
    color: '#374151',
    flex: 1,
  },
  optionTextSelected: {
    color: '#2563eb',
    fontWeight: '500',
  },
});

// Helper: check if template has text questions
function hasTextQuestions(sections: SurveySection[]): boolean {
  return flattenQuestions(sections).some((q) => q.type === 'text');
}

// Helper: build prompt from text answers
function buildTextPrompt(
  sections: SurveySection[],
  answers: Record<string, number | boolean | string>
): string {
  const textQuestions = flattenQuestions(sections).filter((q) => q.type === 'text');
  const parts: string[] = [];
  for (const q of textQuestions) {
    const val = answers[q.id];
    if (typeof val === 'string' && val.trim()) {
      parts.push(`${q.text}: ${val.trim()}`);
    }
  }
  return parts.join('\n');
}

export default function SurveyFormScreen() {
  const { code } = useLocalSearchParams<{ code: string }>();
  const navigation = useNavigation();
  const router = useRouter();
  const scrollRef = useRef<ScrollView>(null);

  const [answers, setAnswers] = useState<Record<string, number | boolean | string>>({});
  const [result, setResult] = useState<SurveyResult | null>(null);
  const [aiRequestText, setAiRequestText] = useState('');
  const [aiAdvice, setAiAdvice] = useState<AIAdviceResult | null>(null);

  const { data: template, isLoading, error } = useQuery<SurveyTemplate>({
    queryKey: ['survey-template', code],
    queryFn: () => surveysApi.getTemplate(code || ''),
    enabled: !!code,
  });

  useEffect(() => {
    if (template) {
      navigation.setOptions({ title: template.name });
      // Initialize answers with default values
      const initialAnswers: Record<string, number | boolean | string> = {};
      const questions = flattenQuestions(template.questions);
      questions.forEach((q) => {
        if (q.type === 'boolean') {
          initialAnswers[q.id] = false;
        } else if (q.type === 'text') {
          initialAnswers[q.id] = '';
        } else if (q.type === 'select' && q.options && q.options.length > 0) {
          // For select, initialize with first option value
          const firstOption = q.options[0];
          initialAnswers[q.id] = typeof firstOption === 'object' ? firstOption.value : 0;
        } else {
          initialAnswers[q.id] = getMin(q);
        }
      });
      setAnswers(initialAnswers);
    }
  }, [template, navigation]);

  const submitMutation = useMutation({
    mutationFn: (surveyAnswers: SurveyAnswer[]) => 
      surveysApi.submitResponse(code || '', surveyAnswers),
    onSuccess: (result, surveyAnswers) => {
      setResult(result);
      setAiAdvice(null);
      // Scroll to top to show result
      setTimeout(() => scrollRef.current?.scrollTo({ y: 0, animated: true }), 100);

      // Auto-trigger AI advice if template has text questions
      if (template && hasTextQuestions(template.questions)) {
        const autoPrompt = buildTextPrompt(template.questions, answers);
        adviceMutation.mutate({ surveyAnswers, text: autoPrompt });
      }
    },
    onError: (err: any) => {
      Alert.alert('Ошибка', err.response?.data?.message || 'Не удалось сохранить');
    },
  });

  const adviceMutation = useMutation({
    mutationFn: (payload: { surveyAnswers: SurveyAnswer[]; text: string }) =>
      surveysApi.createAdvice(code || '', payload.surveyAnswers, payload.text),
    onSuccess: (data) => {
      setAiAdvice(data);
    },
    onError: (err: any) => {
      Alert.alert('Ошибка', err.response?.data?.message || 'Не удалось получить рекомендацию');
    },
  });

  const handleSubmit = () => {
    if (!template) return;

    const questions = flattenQuestions(template.questions);
    const surveyAnswers: SurveyAnswer[] = questions.map((q) => ({
      question_id: q.id,
      value:
        answers[q.id] ??
        (q.type === 'boolean' ? false : q.type === 'text' ? '' : 0),
    }));

    submitMutation.mutate(surveyAnswers);
  };

  const handleGetAdvice = () => {
    if (!template) return;

    const questions = flattenQuestions(template.questions);
    const surveyAnswers: SurveyAnswer[] = questions.map((q) => ({
      question_id: q.id,
      value:
        answers[q.id] ??
        (q.type === 'boolean' ? false : q.type === 'text' ? '' : 0),
    }));

    adviceMutation.mutate({ surveyAnswers, text: aiRequestText });
  };

  const setAnswer = (questionId: string, value: number | boolean | string) => {
    setAnswers((prev) => ({ ...prev, [questionId]: value }));
  };

  if (isLoading) {
    return (
      <View style={styles.loading}>
        <ActivityIndicator size="large" color="#2563eb" />
        <Text style={styles.loadingText}>Загрузка опросника...</Text>
      </View>
    );
  }

  if (error || !template) {
    return (
      <View style={styles.error}>
        <FontAwesome name="exclamation-circle" size={48} color="#dc2626" />
        <Text style={styles.errorText}>Опросник не найден</Text>
        <TouchableOpacity style={styles.backBtn} onPress={() => router.back()}>
          <Text style={styles.backBtnText}>Назад</Text>
        </TouchableOpacity>
      </View>
    );
  }

  return (
    <ScrollView ref={scrollRef} style={styles.container} contentContainerStyle={styles.content}>
      <View style={styles.header}>
        <Text style={styles.title}>{template.name}</Text>
        <Text style={styles.description}>{template.description}</Text>
      </View>

      {result && (
        <View style={styles.resultCard}>
          <Text style={styles.resultTitle}>Результат</Text>
          <Text style={styles.resultValue}>{result.score.toFixed(2)}</Text>
          <Text style={styles.resultInterpretation}>{result.interpretation}</Text>

          <Text style={styles.aiTitle}>AI рекомендации</Text>

          {/* If template has text questions, AI is auto-triggered; otherwise show manual input */}
          {hasTextQuestions(template.questions) ? (
            adviceMutation.isPending ? (
              <View style={styles.aiLoadingBlock}>
                <ActivityIndicator size="small" color="#2563eb" />
                <Text style={styles.aiLoadingText}>Анализируем данные...</Text>
              </View>
            ) : aiAdvice ? (
              <View style={styles.aiAnswerBlock}>
                <Text style={styles.aiAnswerText}>{aiAdvice.advice_text}</Text>
                <Text style={styles.aiDisclaimer}>{aiAdvice.disclaimer}</Text>
                <TouchableOpacity style={styles.aiHistoryBtn} onPress={() => router.push('/(tabs)/advice' as any)}>
                  <Text style={styles.aiHistoryBtnText}>Открыть историю рекомендаций</Text>
                </TouchableOpacity>
              </View>
            ) : null
          ) : (
            <>
              <Text style={styles.aiHint}>
                Можно добавить комментарий (необязательно) и получить пояснение результатов.
              </Text>
              <TextInput
                style={styles.aiInput}
                placeholder="Опишите самочувствие/жалобы (необязательно)"
                placeholderTextColor="#9ca3af"
                value={aiRequestText}
                onChangeText={setAiRequestText}
                multiline
              />

              <TouchableOpacity
                style={[styles.aiButton, adviceMutation.isPending && styles.submitButtonDisabled]}
                onPress={handleGetAdvice}
                disabled={adviceMutation.isPending}
              >
                {adviceMutation.isPending ? (
                  <ActivityIndicator color="#fff" />
                ) : (
                  <>
                    <FontAwesome name="comment" size={18} color="#fff" style={{ marginRight: 8 }} />
                    <Text style={styles.submitButtonText}>Получить рекомендацию</Text>
                  </>
                )}
              </TouchableOpacity>

              {aiAdvice && (
                <View style={styles.aiAnswerBlock}>
                  <Text style={styles.aiAnswerText}>{aiAdvice.advice_text}</Text>
                  <Text style={styles.aiDisclaimer}>{aiAdvice.disclaimer}</Text>
                  <TouchableOpacity style={styles.aiHistoryBtn} onPress={() => router.push('/(tabs)/advice' as any)}>
                    <Text style={styles.aiHistoryBtnText}>Открыть историю рекомендаций</Text>
                  </TouchableOpacity>
                </View>
              )}
            </>
          )}

          <TouchableOpacity style={styles.backBtn} onPress={() => router.back()}>
            <Text style={styles.backBtnText}>Готово</Text>
          </TouchableOpacity>
        </View>
      )}

      <View style={styles.questions}>
        {template.questions.map((section, sectionIndex) => (
          <View key={`${section.section}-${sectionIndex}`} style={styles.sectionBlock}>
            {!!section.title && <Text style={styles.sectionTitle}>{section.title}</Text>}
            {section.questions.map((question, index) => {
              const globalIndex =
                template.questions
                  .slice(0, sectionIndex)
                  .reduce((acc, s) => acc + (s.questions?.length || 0), 0) +
                index;

              return (
                <View key={`${section.section}-${question.id}`} style={styles.questionCard}>
                  <Text style={styles.questionNumber}>Вопрос {globalIndex + 1}</Text>
                  <Text style={styles.questionText}>{question.text}</Text>

                  {question.type === 'boolean' && (
                    <View style={styles.booleanAnswer}>
                      <Text style={styles.booleanLabel}>
                        {answers[question.id] ? 'Да' : 'Нет'}
                      </Text>
                      <Switch
                        value={Boolean(answers[question.id])}
                        onValueChange={(val) => setAnswer(question.id, val)}
                        trackColor={{ false: '#e5e7eb', true: '#93c5fd' }}
                        thumbColor={answers[question.id] ? '#2563eb' : '#9ca3af'}
                      />
                    </View>
                  )}

                  {(question.type === 'scale' || question.type === 'vas' || question.type === 'vas100') && (
                    <View style={styles.vasAnswer}>
                      <View style={styles.vasLabels}>
                        <Text style={styles.vasMinLabel}>{getMin(question)}</Text>
                        <Text style={styles.vasValue}>
                          {typeof answers[question.id] === 'number'
                            ? (answers[question.id] as number).toFixed(1)
                            : '0.0'}
                        </Text>
                        <Text style={styles.vasMaxLabel}>{getMax(question)}</Text>
                      </View>
                      <Slider
                        style={styles.slider}
                        minimumValue={getMin(question)}
                        maximumValue={getMax(question)}
                        step={question.type === 'scale' ? 1 : 0.5}
                        value={typeof answers[question.id] === 'number' ? (answers[question.id] as number) : getMin(question)}
                        onValueChange={(val) => setAnswer(question.id, val)}
                        minimumTrackTintColor="#2563eb"
                        maximumTrackTintColor="#e5e7eb"
                        thumbTintColor="#2563eb"
                      />
                    </View>
                  )}

                  {question.type === 'number' && (
                    <View style={styles.numberAnswer}>
                      <TextInput
                        style={styles.numberInput}
                        placeholder="0"
                        placeholderTextColor="#9ca3af"
                        keyboardType="numeric"
                        value={
                          typeof answers[question.id] === 'number'
                            ? String(answers[question.id])
                            : ''
                        }
                        onChangeText={(txt) => {
                          const normalized = txt.replace(',', '.').trim();
                          const n = normalized === '' ? getMin(question) : Number(normalized);
                          setAnswer(question.id, Number.isFinite(n) ? n : getMin(question));
                        }}
                      />
                    </View>
                  )}

                  {question.type === 'text' && (
                    <View style={styles.textAnswer}>
                      <TextInput
                        style={styles.textInput}
                        placeholder={question.placeholder || ''}
                        placeholderTextColor="#9ca3af"
                        value={typeof answers[question.id] === 'string' ? (answers[question.id] as string) : ''}
                        onChangeText={(txt) => setAnswer(question.id, txt)}
                        multiline
                        textAlignVertical="top"
                      />
                    </View>
                  )}

                  {question.type === 'select' && question.options && (
                    <SelectInput
                      options={question.options}
                      value={answers[question.id] as number}
                      onChange={(val) => setAnswer(question.id, val)}
                    />
                  )}
                </View>
              );
            })}
          </View>
        ))}
      </View>

      <TouchableOpacity
        style={[styles.submitButton, submitMutation.isPending && styles.submitButtonDisabled]}
        onPress={handleSubmit}
        disabled={submitMutation.isPending}
      >
        {submitMutation.isPending ? (
          <ActivityIndicator color="#fff" />
        ) : (
          <>
            <FontAwesome name="check" size={18} color="#fff" style={{ marginRight: 8 }} />
            <Text style={styles.submitButtonText}>Завершить опрос</Text>
          </>
        )}
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f3f4f6',
  },
  content: {
    padding: 16,
    paddingBottom: 32,
  },
  loading: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f3f4f6',
  },
  loadingText: {
    marginTop: 12,
    color: '#6b7280',
    fontSize: 16,
  },
  error: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f3f4f6',
  },
  errorText: {
    marginTop: 12,
    color: '#dc2626',
    fontSize: 18,
    fontWeight: '600',
  },
  backBtn: {
    marginTop: 24,
    backgroundColor: '#2563eb',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  backBtnText: {
    color: '#fff',
    fontWeight: '600',
  },
  header: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 20,
    marginBottom: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  resultCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
  },
  resultTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#111827',
    marginBottom: 8,
  },
  resultValue: {
    fontSize: 32,
    fontWeight: '800',
    color: '#111827',
  },
  resultInterpretation: {
    marginTop: 8,
    fontSize: 14,
    color: '#374151',
  },
  aiTitle: {
    marginTop: 16,
    fontSize: 16,
    fontWeight: '700',
    color: '#111827',
  },
  aiHint: {
    marginTop: 6,
    fontSize: 12,
    color: '#6b7280',
  },
  aiInput: {
    marginTop: 10,
    borderWidth: 1,
    borderColor: '#e5e7eb',
    borderRadius: 10,
    padding: 12,
    minHeight: 80,
    color: '#111827',
    backgroundColor: '#fff',
    textAlignVertical: 'top',
  },
  aiButton: {
    marginTop: 12,
    backgroundColor: '#2563eb',
    paddingVertical: 14,
    borderRadius: 12,
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
  },
  aiAnswerBlock: {
    marginTop: 14,
    borderTopWidth: 1,
    borderTopColor: '#e5e7eb',
    paddingTop: 14,
  },
  aiAnswerText: {
    fontSize: 14,
    color: '#111827',
    lineHeight: 20,
  },
  aiDisclaimer: {
    marginTop: 10,
    fontSize: 12,
    color: '#6b7280',
    lineHeight: 16,
  },
  aiHistoryBtn: {
    marginTop: 12,
    paddingVertical: 10,
  },
  aiHistoryBtnText: {
    color: '#2563eb',
    fontWeight: '600',
  },
  title: {
    fontSize: 22,
    fontWeight: 'bold',
    color: '#1f2937',
  },
  description: {
    fontSize: 14,
    color: '#6b7280',
    marginTop: 8,
    lineHeight: 20,
  },
  questions: {
    gap: 12,
  },
  sectionBlock: {
    gap: 12,
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: '700',
    color: '#111827',
    marginTop: 4,
    marginBottom: 4,
    paddingHorizontal: 4,
  },
  questionCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  questionNumber: {
    fontSize: 12,
    color: '#2563eb',
    fontWeight: '600',
    marginBottom: 4,
  },
  numberAnswer: {
    marginTop: 4,
  },
  numberInput: {
    borderWidth: 1,
    borderColor: '#e5e7eb',
    borderRadius: 10,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
    color: '#111827',
    backgroundColor: '#fff',
  },
  textAnswer: {
    marginTop: 4,
  },
  textInput: {
    borderWidth: 1,
    borderColor: '#e5e7eb',
    borderRadius: 10,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
    color: '#111827',
    backgroundColor: '#fff',
    minHeight: 96,
  },
  questionText: {
    fontSize: 16,
    color: '#1f2937',
    fontWeight: '500',
    marginBottom: 16,
    lineHeight: 22,
  },
  booleanAnswer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 8,
  },
  booleanLabel: {
    fontSize: 16,
    color: '#374151',
    fontWeight: '500',
  },
  vasAnswer: {
    paddingVertical: 8,
  },
  vasLabels: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  vasMinLabel: {
    fontSize: 14,
    color: '#6b7280',
  },
  vasValue: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#2563eb',
  },
  vasMaxLabel: {
    fontSize: 14,
    color: '#6b7280',
  },
  slider: {
    width: '100%',
    height: 40,
  },
  submitButton: {
    backgroundColor: '#2563eb',
    borderRadius: 12,
    padding: 16,
    marginTop: 24,
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
  },
  submitButtonDisabled: {
    backgroundColor: '#93c5fd',
  },
  submitButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
  },
  aiLoadingBlock: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    gap: 10,
  },
  aiLoadingText: {
    fontSize: 15,
    color: '#6b7280',
  },
});
