import { View, Text, StyleSheet, ScrollView, TouchableOpacity, ActivityIndicator } from 'react-native';
import { useRouter } from 'expo-router';
import { useQuery } from '@tanstack/react-query';
import FontAwesome from '@expo/vector-icons/FontAwesome';
import { surveysApi } from '../../src/api/client';
import { useAuthStore } from '../../src/store/auth';

export default function HomeScreen() {
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const authLoading = useAuthStore((s) => s.isLoading);

  const { data: templates, isLoading: loadingTemplates } = useQuery({
    queryKey: ['survey-templates'],
    queryFn: () => surveysApi.getTemplates(),
    enabled: isAuthenticated,
  });

  if (authLoading) {
    return (
      <View style={styles.container}>
        <ActivityIndicator size="large" color="#2563eb" />
      </View>
    );
  }

  if (!isAuthenticated) {
    return (
      <View style={styles.container}>
        <View style={styles.welcomeCard}>
          <FontAwesome name="stethoscope" size={64} color="#2563eb" />
          <Text style={styles.welcomeTitle}>Предоперационная оценка</Text>
          <Text style={styles.welcomeSubtitle}>
            Шкалы оценки риска периоперационных осложнений
          </Text>
          <TouchableOpacity
            style={styles.authButton}
            onPress={() => router.push('/auth/login')}
          >
            <Text style={styles.authButtonText}>Войти</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={styles.authButtonSecondary}
            onPress={() => router.push('/auth/register')}
          >
            <Text style={styles.authButtonSecondaryText}>Регистрация</Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  return (
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.greeting}>
          Здравствуйте, {user?.first_name || 'Пользователь'}!
        </Text>
        <Text style={styles.date}>
          {new Date().toLocaleDateString('ru-RU', {
            weekday: 'long',
            day: 'numeric',
            month: 'long',
          })}
        </Text>
      </View>

      {/* Info Card */}
      <View style={styles.section}>
        <View style={styles.infoCard}>
          <FontAwesome name="info-circle" size={24} color="#2563eb" />
          <Text style={styles.infoText}>
            Выберите шкалу для оценки периоперационного риска. 
            Результат поможет в планировании предоперационной подготовки.
          </Text>
        </View>
      </View>

      {/* Available Scales */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Шкалы оценки риска</Text>
        {loadingTemplates ? (
          <ActivityIndicator size="small" color="#2563eb" />
        ) : templates?.length ? (
          templates.map((t: any) => (
            <TouchableOpacity
              key={t.id}
              style={styles.listItem}
              onPress={() => router.push(`/surveys/${t.code}`)}
            >
              <View style={styles.listItemContent}>
                <View style={styles.iconContainer}>
                  <FontAwesome 
                    name={getScaleIcon(t.code)} 
                    size={24} 
                    color={getScaleColor(t.code)} 
                  />
                </View>
                <View style={styles.listItemText}>
                  <Text style={styles.listItemTitle}>{t.name}</Text>
                  <Text style={styles.listItemSubtitle}>{getScaleDescription(t.code)}</Text>
                </View>
              </View>
              <FontAwesome name="chevron-right" size={16} color="#9ca3af" />
            </TouchableOpacity>
          ))
        ) : (
          <Text style={styles.emptyText}>Нет доступных опросников</Text>
        )}
      </View>
    </ScrollView>
  );
}

function getScaleIcon(code: string): React.ComponentProps<typeof FontAwesome>['name'] {
  switch (code) {
    case 'ASA': return 'user-md';
    case 'RCRI': return 'heartbeat';
    case 'GOLDMAN': return 'heart';
    case 'CAPRINI': return 'tint';
    default: return 'list-alt';
  }
}

function getScaleColor(code: string): string {
  switch (code) {
    case 'ASA': return '#2563eb';
    case 'RCRI': return '#dc2626';
    case 'GOLDMAN': return '#ea580c';
    case 'CAPRINI': return '#7c3aed';
    default: return '#6b7280';
  }
}

function getScaleDescription(code: string): string {
  switch (code) {
    case 'ASA': return 'Общая оценка физического статуса';
    case 'RCRI': return 'Кардиальный риск (индекс Ли)';
    case 'GOLDMAN': return 'Оригинальный кардиальный индекс';
    case 'CAPRINI': return 'Риск венозных тромбоэмболий';
    default: return '';
  }
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f3f4f6',
  },
  welcomeCard: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  welcomeTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#1f2937',
    marginTop: 24,
    textAlign: 'center',
  },
  welcomeSubtitle: {
    fontSize: 16,
    color: '#6b7280',
    marginTop: 8,
    textAlign: 'center',
  },
  authButton: {
    backgroundColor: '#2563eb',
    paddingHorizontal: 48,
    paddingVertical: 16,
    borderRadius: 12,
    marginTop: 32,
    width: '100%',
  },
  authButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
    textAlign: 'center',
  },
  authButtonSecondary: {
    paddingHorizontal: 48,
    paddingVertical: 16,
    marginTop: 12,
  },
  authButtonSecondaryText: {
    color: '#2563eb',
    fontSize: 16,
    fontWeight: '500',
    textAlign: 'center',
  },
  header: {
    backgroundColor: '#2563eb',
    padding: 24,
    paddingTop: 60,
  },
  greeting: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
  },
  date: {
    fontSize: 14,
    color: '#bfdbfe',
    marginTop: 4,
  },
  section: {
    padding: 16,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#374151',
    marginBottom: 12,
  },
  infoCard: {
    backgroundColor: '#eff6ff',
    padding: 16,
    borderRadius: 12,
    flexDirection: 'row',
    alignItems: 'flex-start',
    gap: 12,
  },
  infoText: {
    flex: 1,
    fontSize: 14,
    color: '#1e40af',
    lineHeight: 20,
  },
  listItem: {
    backgroundColor: '#fff',
    padding: 16,
    borderRadius: 12,
    marginBottom: 8,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  listItemContent: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
  },
  iconContainer: {
    width: 48,
    height: 48,
    borderRadius: 12,
    backgroundColor: '#f3f4f6',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  listItemText: {
    flex: 1,
  },
  listItemTitle: {
    fontSize: 16,
    fontWeight: '500',
    color: '#1f2937',
  },
  listItemSubtitle: {
    fontSize: 14,
    color: '#6b7280',
    marginTop: 2,
  },
  emptyText: {
    textAlign: 'center',
    color: '#9ca3af',
    padding: 16,
  },
});
