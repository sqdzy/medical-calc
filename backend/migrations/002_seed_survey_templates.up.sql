-- 002_seed_survey_templates.up.sql
-- Seed BVAS and other medical survey templates

-- BVAS v3 (Birmingham Vasculitis Activity Score)
INSERT INTO survey_templates (id, code, name, description, category, questions, scoring_logic, interpretation_rules, version) VALUES
(
    '00000000-0000-0000-0000-000000000001',
    'BVAS_V3',
    'Birmingham Vasculitis Activity Score v3',
    'Индекс активности васкулита BVAS v3 для оценки системных васкулитов',
    'vasculitis',
    '[
        {
            "section": "general",
            "title": "Общие симптомы",
            "questions": [
                {"id": "gen_1", "text": "Недомогание", "type": "boolean", "score": 1},
                {"id": "gen_2", "text": "Миалгия", "type": "boolean", "score": 1},
                {"id": "gen_3", "text": "Артралгия/артрит", "type": "boolean", "score": 1},
                {"id": "gen_4", "text": "Лихорадка (≥38°C)", "type": "boolean", "score": 2},
                {"id": "gen_5", "text": "Потеря веса (≥2 кг)", "type": "boolean", "score": 2}
            ]
        },
        {
            "section": "cutaneous",
            "title": "Кожные проявления",
            "questions": [
                {"id": "cut_1", "text": "Инфаркт", "type": "boolean", "score": 2},
                {"id": "cut_2", "text": "Пурпура", "type": "boolean", "score": 2},
                {"id": "cut_3", "text": "Язвы", "type": "boolean", "score": 4},
                {"id": "cut_4", "text": "Гангрена", "type": "boolean", "score": 6},
                {"id": "cut_5", "text": "Другие кожные васкулиты", "type": "boolean", "score": 2}
            ]
        },
        {
            "section": "mucous_membranes",
            "title": "Слизистые оболочки / глаза",
            "questions": [
                {"id": "muc_1", "text": "Язвы во рту", "type": "boolean", "score": 1},
                {"id": "muc_2", "text": "Язвы гениталий", "type": "boolean", "score": 1},
                {"id": "muc_3", "text": "Конъюнктивит", "type": "boolean", "score": 1},
                {"id": "muc_4", "text": "Эписклерит/склерит", "type": "boolean", "score": 2},
                {"id": "muc_5", "text": "Увеит", "type": "boolean", "score": 6},
                {"id": "muc_6", "text": "Внезапная потеря зрения", "type": "boolean", "score": 6}
            ]
        },
        {
            "section": "ent",
            "title": "ЛОР-органы",
            "questions": [
                {"id": "ent_1", "text": "Кровянистые выделения из носа", "type": "boolean", "score": 2},
                {"id": "ent_2", "text": "Синусит", "type": "boolean", "score": 2},
                {"id": "ent_3", "text": "Подсвязочный стеноз", "type": "boolean", "score": 6},
                {"id": "ent_4", "text": "Кондуктивная тугоухость", "type": "boolean", "score": 3},
                {"id": "ent_5", "text": "Сенсоневральная тугоухость", "type": "boolean", "score": 6}
            ]
        },
        {
            "section": "chest",
            "title": "Грудная клетка",
            "questions": [
                {"id": "che_1", "text": "Хрипы", "type": "boolean", "score": 2},
                {"id": "che_2", "text": "Узелки/полости", "type": "boolean", "score": 3},
                {"id": "che_3", "text": "Плеврит/плевральный выпот", "type": "boolean", "score": 4},
                {"id": "che_4", "text": "Инфильтраты", "type": "boolean", "score": 4},
                {"id": "che_5", "text": "Эндобронхиальное поражение", "type": "boolean", "score": 4},
                {"id": "che_6", "text": "Массивное кровохарканье / альвеолярное кровотечение", "type": "boolean", "score": 6},
                {"id": "che_7", "text": "Дыхательная недостаточность", "type": "boolean", "score": 6}
            ]
        },
        {
            "section": "cardiovascular",
            "title": "Сердечно-сосудистая система",
            "questions": [
                {"id": "cvs_1", "text": "Потеря пульса", "type": "boolean", "score": 4},
                {"id": "cvs_2", "text": "Аортальная недостаточность", "type": "boolean", "score": 4},
                {"id": "cvs_3", "text": "Перикардит", "type": "boolean", "score": 3},
                {"id": "cvs_4", "text": "Кардиомиопатия/сердечная недостаточность", "type": "boolean", "score": 6}
            ]
        },
        {
            "section": "abdominal",
            "title": "Абдоминальные проявления",
            "questions": [
                {"id": "abd_1", "text": "Перитонит", "type": "boolean", "score": 9},
                {"id": "abd_2", "text": "Кровавая диарея", "type": "boolean", "score": 6},
                {"id": "abd_3", "text": "Ишемическая боль в животе", "type": "boolean", "score": 6}
            ]
        },
        {
            "section": "renal",
            "title": "Почки",
            "questions": [
                {"id": "ren_1", "text": "Артериальная гипертензия", "type": "boolean", "score": 4},
                {"id": "ren_2", "text": "Протеинурия (>1+ или >0.2 г/сут)", "type": "boolean", "score": 4},
                {"id": "ren_3", "text": "Гематурия (≥10 эритроцитов/поле зрения)", "type": "boolean", "score": 6},
                {"id": "ren_4", "text": "Креатинин 125-249 мкмоль/л", "type": "boolean", "score": 6},
                {"id": "ren_5", "text": "Креатинин 250-499 мкмоль/л", "type": "boolean", "score": 8},
                {"id": "ren_6", "text": "Креатинин ≥500 мкмоль/л", "type": "boolean", "score": 10},
                {"id": "ren_7", "text": "Повышение креатинина >30% или снижение СКФ >25%", "type": "boolean", "score": 6}
            ]
        },
        {
            "section": "nervous",
            "title": "Нервная система",
            "questions": [
                {"id": "ner_1", "text": "Головная боль", "type": "boolean", "score": 1},
                {"id": "ner_2", "text": "Менингит", "type": "boolean", "score": 3},
                {"id": "ner_3", "text": "Органическая спутанность сознания", "type": "boolean", "score": 3},
                {"id": "ner_4", "text": "Судороги (не гипертензивные)", "type": "boolean", "score": 9},
                {"id": "ner_5", "text": "Инсульт", "type": "boolean", "score": 9},
                {"id": "ner_6", "text": "Поражение спинного мозга", "type": "boolean", "score": 9},
                {"id": "ner_7", "text": "Поражение черепных нервов", "type": "boolean", "score": 6},
                {"id": "ner_8", "text": "Сенсорная периферическая нейропатия", "type": "boolean", "score": 6},
                {"id": "ner_9", "text": "Моторная мононейропатия", "type": "boolean", "score": 9}
            ]
        }
    ]'::jsonb,
    '{
        "type": "sum",
        "sections": ["general", "cutaneous", "mucous_membranes", "ent", "chest", "cardiovascular", "abdominal", "renal", "nervous"]
    }'::jsonb,
    '{
        "ranges": [
            {"min": 0, "max": 0, "category": "remission", "description": "Ремиссия"},
            {"min": 1, "max": 5, "category": "low", "description": "Низкая активность"},
            {"min": 6, "max": 15, "category": "moderate", "description": "Умеренная активность"},
            {"min": 16, "max": 999, "category": "high", "description": "Высокая активность"}
        ]
    }'::jsonb,
    1
);

-- DAS28 (Disease Activity Score 28)
INSERT INTO survey_templates (id, code, name, description, category, questions, scoring_logic, interpretation_rules, version) VALUES
(
    '00000000-0000-0000-0000-000000000002',
    'DAS28_CRP',
    'Disease Activity Score 28 (CRP)',
    'Индекс активности ревматоидного артрита DAS28 с использованием СРБ',
    'arthritis',
    '[
        {
            "section": "joints",
            "title": "Оценка суставов",
            "questions": [
                {"id": "tjc28", "text": "Число болезненных суставов (0-28)", "type": "number", "min": 0, "max": 28},
                {"id": "sjc28", "text": "Число припухших суставов (0-28)", "type": "number", "min": 0, "max": 28}
            ]
        },
        {
            "section": "lab",
            "title": "Лабораторные данные",
            "questions": [
                {"id": "crp", "text": "СРБ (мг/л)", "type": "number", "min": 0, "max": 300}
            ]
        },
        {
            "section": "patient",
            "title": "Оценка пациента",
            "questions": [
                {"id": "gh", "text": "Общая оценка здоровья пациентом (0-100 мм по ВАШ)", "type": "number", "min": 0, "max": 100}
            ]
        }
    ]'::jsonb,
    '{
        "type": "formula",
        "formula": "0.56*sqrt(tjc28) + 0.28*sqrt(sjc28) + 0.36*ln(crp+1) + 0.014*gh + 0.96"
    }'::jsonb,
    '{
        "ranges": [
            {"min": 0, "max": 2.6, "category": "remission", "description": "Ремиссия"},
            {"min": 2.6, "max": 3.2, "category": "low", "description": "Низкая активность"},
            {"min": 3.2, "max": 5.1, "category": "moderate", "description": "Умеренная активность"},
            {"min": 5.1, "max": 10, "category": "high", "description": "Высокая активность"}
        ]
    }'::jsonb,
    1
);

-- BASDAI (Bath Ankylosing Spondylitis Disease Activity Index)
INSERT INTO survey_templates (id, code, name, description, category, questions, scoring_logic, interpretation_rules, version) VALUES
(
    '00000000-0000-0000-0000-000000000003',
    'BASDAI',
    'Bath Ankylosing Spondylitis Disease Activity Index',
    'Индекс активности анкилозирующего спондилита BASDAI',
    'spondylitis',
    '[
        {
            "section": "symptoms",
            "title": "Оценка симптомов за последнюю неделю",
            "questions": [
                {"id": "q1", "text": "Как бы Вы оценили общий уровень усталости/утомляемости?", "type": "scale", "min": 0, "max": 10, "labels": {"0": "Нет", "10": "Очень сильная"}},
                {"id": "q2", "text": "Как бы Вы оценили боль в шее, спине или тазобедренных суставах?", "type": "scale", "min": 0, "max": 10, "labels": {"0": "Нет", "10": "Очень сильная"}},
                {"id": "q3", "text": "Как бы Вы оценили боль/припухлость в других суставах (кроме шеи, спины, тазобедренных)?", "type": "scale", "min": 0, "max": 10, "labels": {"0": "Нет", "10": "Очень сильная"}},
                {"id": "q4", "text": "Как бы Вы оценили дискомфорт при прикосновении или надавливании?", "type": "scale", "min": 0, "max": 10, "labels": {"0": "Нет", "10": "Очень сильный"}},
                {"id": "q5", "text": "Как бы Вы оценили выраженность утренней скованности с момента пробуждения?", "type": "scale", "min": 0, "max": 10, "labels": {"0": "Нет", "10": "Очень сильная"}},
                {"id": "q6", "text": "Как долго длится утренняя скованность с момента пробуждения?", "type": "scale", "min": 0, "max": 10, "labels": {"0": "0 часов", "5": "1 час", "10": "2+ часов"}}
            ]
        }
    ]'::jsonb,
    '{
        "type": "formula",
        "formula": "(q1 + q2 + q3 + q4 + (q5 + q6) / 2) / 5"
    }'::jsonb,
    '{
        "ranges": [
            {"min": 0, "max": 4, "category": "low", "description": "Низкая активность"},
            {"min": 4, "max": 10, "category": "high", "description": "Высокая активность (показано назначение биологической терапии)"}
        ]
    }'::jsonb,
    1
);
