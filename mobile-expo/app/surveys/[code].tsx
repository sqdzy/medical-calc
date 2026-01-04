import { useState, useEffect } from 'react';
import {
  View,
  Text,
  ScrollView,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  Switch,
  Alert,
} from 'react-native';
import { useLocalSearchParams, useNavigation, useRouter } from 'expo-router';
import { useQuery, useMutation } from '@tanstack/react-query';
import Slider from '@react-native-community/slider';
import FontAwesome from '@expo/vector-icons/FontAwesome';
import { surveysApi } from '../../src/api/client';
import type { SurveyTemplate, SurveyQuestion, SurveyAnswer } from '../../src/types';

export default function SurveyFormScreen() {
  const { code } = useLocalSearchParams<{ code: string }>();
  const navigation = useNavigation();
  const router = useRouter();

  const [answers, setAnswers] = useState<Record<string, number | boolean>>({});

  const { data: template, isLoading, error } = useQuery<SurveyTemplate>({
    queryKey: ['survey-template', code],
    queryFn: () => surveysApi.getTemplate(code || ''),
    enabled: !!code,
  });

  useEffect(() => {
    if (template) {
      navigation.setOptions({ title: template.name });
      // Initialize answers with default values
      const initialAnswers: Record<string, number | boolean> = {};
      template.questions.forEach((q) => {
        if (q.type === 'boolean') {
          initialAnswers[q.id] = false;
        } else if (q.type === 'vas') {
          initialAnswers[q.id] = q.min_value || 0;
        }
      });
      setAnswers(initialAnswers);
    }
  }, [template, navigation]);

  const submitMutation = useMutation({
    mutationFn: (surveyAnswers: SurveyAnswer[]) => 
      surveysApi.submitResponse(code || '', surveyAnswers),
    onSuccess: (result) => {
      Alert.alert(
        'Результат',
        `${template?.name}: ${result.score.toFixed(2)}\n\n${result.interpretation}`,
        [
          {
            text: 'OK',
            onPress: () => router.back(),
          },
        ]
      );
    },
    onError: (err: any) => {
      Alert.alert('Ошибка', err.response?.data?.message || 'Не удалось сохранить');
    },
  });

  const handleSubmit = () => {
    if (!template) return;

    const surveyAnswers: SurveyAnswer[] = template.questions.map((q) => ({
      question_id: q.id,
      value: answers[q.id] ?? (q.type === 'boolean' ? false : 0),
    }));

    submitMutation.mutate(surveyAnswers);
  };

  const setAnswer = (questionId: string, value: number | boolean) => {
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
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <View style={styles.header}>
        <Text style={styles.title}>{template.name}</Text>
        <Text style={styles.description}>{template.description}</Text>
      </View>

      <View style={styles.questions}>
        {template.questions.map((question, index) => (
          <View key={question.id} style={styles.questionCard}>
            <Text style={styles.questionNumber}>Вопрос {index + 1}</Text>
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

            {question.type === 'vas' && (
              <View style={styles.vasAnswer}>
                <View style={styles.vasLabels}>
                  <Text style={styles.vasMinLabel}>{question.min_value || 0}</Text>
                  <Text style={styles.vasValue}>
                    {typeof answers[question.id] === 'number'
                      ? (answers[question.id] as number).toFixed(1)
                      : '0.0'}
                  </Text>
                  <Text style={styles.vasMaxLabel}>{question.max_value || 10}</Text>
                </View>
                <Slider
                  style={styles.slider}
                  minimumValue={question.min_value || 0}
                  maximumValue={question.max_value || 10}
                  step={0.5}
                  value={typeof answers[question.id] === 'number' ? answers[question.id] as number : 0}
                  onValueChange={(val) => setAnswer(question.id, val)}
                  minimumTrackTintColor="#2563eb"
                  maximumTrackTintColor="#e5e7eb"
                  thumbTintColor="#2563eb"
                />
              </View>
            )}
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
});
