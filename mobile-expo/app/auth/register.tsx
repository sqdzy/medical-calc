import { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import FontAwesome from '@expo/vector-icons/FontAwesome';
import { useAuthStore } from '../../src/store/auth';

const registerSchema = z.object({
  email: z.string().email('Некорректный email'),
  password: z
    .string()
    .min(8, 'Минимум 8 символов')
    .regex(/[A-Z]/, 'Добавьте хотя бы 1 заглавную букву')
    .regex(/[a-z]/, 'Добавьте хотя бы 1 строчную букву')
    .regex(/[0-9]/, 'Добавьте хотя бы 1 цифру'),
  first_name: z.string().min(2, 'Минимум 2 символа'),
  last_name: z.string().min(2, 'Минимум 2 символа'),
});

type RegisterForm = z.infer<typeof registerSchema>;

export default function RegisterScreen() {
  const router = useRouter();
  const register = useAuthStore((s) => s.register);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
    defaultValues: { email: '', password: '', first_name: '', last_name: '' },
  });

  const onSubmit = async (data: RegisterForm) => {
    setIsLoading(true);
    setError(null);
    try {
      await register(data);
      router.replace('/(tabs)');
    } catch (err: any) {
      setError(err.response?.data?.message || 'Ошибка регистрации');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView contentContainerStyle={styles.scroll}>
        <TouchableOpacity style={styles.backButton} onPress={() => router.back()}>
          <FontAwesome name="arrow-left" size={20} color="#374151" />
        </TouchableOpacity>

        <View style={styles.header}>
          <FontAwesome name="user-plus" size={48} color="#2563eb" />
          <Text style={styles.title}>Регистрация</Text>
          <Text style={styles.subtitle}>Создайте новый аккаунт</Text>
        </View>

        {error && (
          <View style={styles.errorBox}>
            <Text style={styles.errorText}>{error}</Text>
          </View>
        )}

        <View style={styles.form}>
          <View style={styles.row}>
            <View style={styles.halfField}>
              <Text style={styles.label}>Имя</Text>
              <Controller
                control={control}
                name="first_name"
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    style={[styles.input, errors.first_name && styles.inputError]}
                    placeholder="Иван"
                    placeholderTextColor="#9ca3af"
                    onBlur={onBlur}
                    onChangeText={onChange}
                    value={value}
                  />
                )}
              />
              {errors.first_name && (
                <Text style={styles.fieldError}>{errors.first_name.message}</Text>
              )}
            </View>
            <View style={styles.halfField}>
              <Text style={styles.label}>Фамилия</Text>
              <Controller
                control={control}
                name="last_name"
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    style={[styles.input, errors.last_name && styles.inputError]}
                    placeholder="Иванов"
                    placeholderTextColor="#9ca3af"
                    onBlur={onBlur}
                    onChangeText={onChange}
                    value={value}
                  />
                )}
              />
              {errors.last_name && (
                <Text style={styles.fieldError}>{errors.last_name.message}</Text>
              )}
            </View>
          </View>

          <Text style={styles.label}>Email</Text>
          <Controller
            control={control}
            name="email"
            render={({ field: { onChange, onBlur, value } }) => (
              <TextInput
                style={[styles.input, errors.email && styles.inputError]}
                placeholder="email@example.com"
                placeholderTextColor="#9ca3af"
                keyboardType="email-address"
                autoCapitalize="none"
                onBlur={onBlur}
                onChangeText={onChange}
                value={value}
              />
            )}
          />
          {errors.email && <Text style={styles.fieldError}>{errors.email.message}</Text>}

          <Text style={styles.label}>Пароль</Text>
          <Controller
            control={control}
            name="password"
            render={({ field: { onChange, onBlur, value } }) => (
              <TextInput
                style={[styles.input, errors.password && styles.inputError]}
                placeholder="••••••••"
                placeholderTextColor="#9ca3af"
                secureTextEntry
                onBlur={onBlur}
                onChangeText={onChange}
                value={value}
              />
            )}
          />
          <Text style={styles.passwordHint}>
            Пароль должен быть минимум 8 символов и содержать заглавную букву, строчную букву и цифру.
          </Text>
          {errors.password && <Text style={styles.fieldError}>{errors.password.message}</Text>}

          <TouchableOpacity
            style={styles.submitButton}
            onPress={handleSubmit(onSubmit)}
            disabled={isLoading}
          >
            {isLoading ? (
              <ActivityIndicator color="#fff" />
            ) : (
              <Text style={styles.submitButtonText}>Зарегистрироваться</Text>
            )}
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.linkButton}
            onPress={() => router.push('/auth/login')}
          >
            <Text style={styles.linkText}>
              Уже есть аккаунт? <Text style={styles.linkTextBold}>Войти</Text>
            </Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
  },
  scroll: {
    flexGrow: 1,
    padding: 24,
    paddingTop: 60,
  },
  backButton: {
    width: 40,
    height: 40,
    justifyContent: 'center',
    alignItems: 'center',
  },
  header: {
    alignItems: 'center',
    marginTop: 16,
    marginBottom: 24,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#1f2937',
    marginTop: 16,
  },
  subtitle: {
    fontSize: 16,
    color: '#6b7280',
    marginTop: 4,
  },
  errorBox: {
    backgroundColor: '#fee2e2',
    padding: 12,
    borderRadius: 8,
    marginBottom: 16,
  },
  errorText: {
    color: '#dc2626',
    textAlign: 'center',
  },
  form: {
    flex: 1,
  },
  row: {
    flexDirection: 'row',
    gap: 12,
  },
  halfField: {
    flex: 1,
  },
  label: {
    fontSize: 14,
    fontWeight: '500',
    color: '#374151',
    marginBottom: 6,
    marginTop: 16,
  },
  input: {
    backgroundColor: '#f3f4f6',
    borderRadius: 12,
    padding: 16,
    fontSize: 16,
    borderWidth: 1,
    borderColor: '#e5e7eb',
  },
  inputError: {
    borderColor: '#dc2626',
  },
  fieldError: {
    color: '#dc2626',
    fontSize: 12,
    marginTop: 4,
  },
  passwordHint: {
    color: '#6b7280',
    fontSize: 12,
    marginTop: 6,
    lineHeight: 16,
  },
  submitButton: {
    backgroundColor: '#2563eb',
    borderRadius: 12,
    padding: 16,
    marginTop: 32,
    alignItems: 'center',
  },
  submitButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
  },
  linkButton: {
    marginTop: 24,
    alignItems: 'center',
  },
  linkText: {
    color: '#6b7280',
    fontSize: 14,
  },
  linkTextBold: {
    color: '#2563eb',
    fontWeight: '600',
  },
});
