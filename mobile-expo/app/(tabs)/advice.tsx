import { useMemo } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TouchableOpacity, Alert, FlatList } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useQuery } from '@tanstack/react-query';
import FontAwesome from '@expo/vector-icons/FontAwesome';
import { aiApi } from '../../src/api/client';
import type { AIAdviceResult } from '../../src/types';

function formatRuDate(iso: string) {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString('ru-RU');
}

export default function AdviceHistoryScreen() {
  const { data, isLoading, error, refetch, isFetching } = useQuery<AIAdviceResult[]>({
    queryKey: ['ai-advice-history'],
    queryFn: () => aiApi.listAdvice(50, 0),
  });

  const items = useMemo(() => data || [], [data]);

  if (isLoading) {
    return (
      <SafeAreaView style={styles.loading} edges={['top']}>
        <ActivityIndicator size="large" color="#2563eb" />
        <Text style={styles.loadingText}>Загрузка истории...</Text>
      </SafeAreaView>
    );
  }

  if (error) {
    return (
      <SafeAreaView style={styles.error} edges={['top']}>
        <FontAwesome name="exclamation-circle" size={48} color="#dc2626" />
        <Text style={styles.errorText}>Не удалось загрузить</Text>
        <TouchableOpacity style={styles.retryBtn} onPress={() => refetch()}>
          <Text style={styles.retryBtnText}>Повторить</Text>
        </TouchableOpacity>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container} edges={['top']}>
      <View style={styles.header}>
        <Text style={styles.title}>AI рекомендации</Text>
        <TouchableOpacity style={styles.refreshBtn} onPress={() => refetch()} disabled={isFetching}>
          {isFetching ? (
            <ActivityIndicator color="#2563eb" />
          ) : (
            <FontAwesome name="refresh" size={18} color="#2563eb" />
          )}
        </TouchableOpacity>
      </View>

      {items.length === 0 ? (
        <View style={styles.empty}>
          <Text style={styles.emptyText}>История пока пустая</Text>
        </View>
      ) : (
        <FlatList
          data={items}
          keyExtractor={(it) => it.id}
          contentContainerStyle={styles.list}
          renderItem={({ item }) => (
            <TouchableOpacity
              style={styles.card}
              onPress={() =>
                Alert.alert(
                  `${item.survey_code} • ${formatRuDate(item.created_at)}`,
                  `${item.advice_text}\n\n${item.disclaimer}`,
                  [{ text: 'OK' }]
                )
              }
            >
              <View style={styles.cardTop}>
                <Text style={styles.cardTitle}>{item.survey_code}</Text>
                <Text style={styles.cardDate}>{formatRuDate(item.created_at)}</Text>
              </View>
              {!!item.user_text && <Text style={styles.userText}>Комментарий: {item.user_text}</Text>}
              <Text style={styles.preview} numberOfLines={3}>
                {item.advice_text}
              </Text>
            </TouchableOpacity>
          )}
        />
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f3f4f6',
  },
  header: {
    paddingHorizontal: 16,
    paddingTop: 16,
    paddingBottom: 8,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  title: {
    fontSize: 20,
    fontWeight: '800',
    color: '#111827',
  },
  refreshBtn: {
    padding: 10,
  },
  list: {
    padding: 16,
    paddingBottom: 24,
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 14,
    marginBottom: 12,
  },
  cardTop: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'baseline',
  },
  cardTitle: {
    fontSize: 14,
    fontWeight: '700',
    color: '#111827',
  },
  cardDate: {
    fontSize: 12,
    color: '#6b7280',
  },
  userText: {
    marginTop: 8,
    fontSize: 12,
    color: '#374151',
  },
  preview: {
    marginTop: 8,
    fontSize: 13,
    color: '#111827',
    lineHeight: 18,
  },
  empty: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyText: {
    color: '#6b7280',
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
    fontSize: 16,
  },
  retryBtn: {
    marginTop: 12,
    backgroundColor: '#2563eb',
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 10,
  },
  retryBtnText: {
    color: '#fff',
    fontWeight: '700',
  },
});
