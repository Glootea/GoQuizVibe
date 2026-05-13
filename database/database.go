package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/goquizvibe/config"
	"github.com/goquizvibe/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func Connect(ctx context.Context, c config.Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@localhost:5433/%s?sslmode=disable",
		c.Database.User, c.Database.Password, c.Database.DBName,
	)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return pool, nil
}

func runMigrations() error {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@localhost:5433/%s?sslmode=disable",
		"goquizvibe", "goquizvibe", "goquizvibe",
	)
	m, err := migrate.New("file://sql/migrations", connStr)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("Migrations applied")
	return nil
}

func SeedData(ctx context.Context, pool *pgxpool.Pool) error {
	queries := db.New(pool)

	exists, err := queries.EmailExists(ctx, "teacher@example.com")
	if err == nil && exists {
		return nil
	}

	log.Println("Seeding initial data...")

	hash, _ := bcrypt.GenerateFromPassword([]byte("teacher123"), bcrypt.DefaultCost)
	teacherID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	teacher := db.CreateUserParams{
		ID:           teacherID,
		Name:         "Учитель",
		Email:        "teacher@example.com",
		PasswordHash: string(hash),
		Role:         db.RoleTeacher,
		CreatedAt:    time.Now(),
	}
	if _, err := queries.CreateUser(ctx, teacher); err != nil {
		return fmt.Errorf("failed to create teacher: %w", err)
	}

	hash, _ = bcrypt.GenerateFromPassword([]byte("student123"), bcrypt.DefaultCost)
	studentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	student := db.CreateUserParams{
		ID:           studentID,
		Name:         "Студент",
		Email:        "student@example.com",
		PasswordHash: string(hash),
		Role:         db.RoleStudent,
		CreatedAt:    time.Now(),
	}
	if _, err := queries.CreateUser(ctx, student); err != nil {
		return fmt.Errorf("failed to create student: %w", err)
	}

	quizzes := []struct {
		ID          uuid.UUID
		Title       string
		Description string
		Subject     string
		Grade       int
		TimeLimit   int
		Questions   []struct {
			Text          string
			Type          db.QuestionType
			Options       []string
			CorrectAnswer string
			Explanation   string
			Points        int
		}
	}{
		{
			ID:          uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			Title:       "Математика 5 класс: Дроби и десятичные дроби",
			Description: "Тест на знание обыкновенных и десятичных дробей",
			Subject:     "Математика",
			Grade:       5,
			TimeLimit:   900,
			Questions: []struct {
				Text          string
				Type          db.QuestionType
				Options       []string
				CorrectAnswer string
				Explanation   string
				Points        int
			}{
				{"Чему равна сумма 1/4 + 1/4?", db.QuestionTypeChoice, []string{"1/2", "2/4", "1/8", "1"}, "1/2", "При сложении дробей с одинаковым знаменателем складываем числители: 1+1=2, получаем 2/4 = 1/2", 10},
				{"Переведите десятичную дробь 0.75 в обыкновенную дробь.", db.QuestionTypeOpen, []string{}, "3/4", "0.75 = 75/100 = 3/4 (сокращаем на 25)", 10},
				{"Вычислите: 2.5 + 1.3", db.QuestionTypeChoice, []string{"3.8", "3.7", "3.9", "4.0"}, "3.8", "Десятые доли: 0.5 + 0.3 = 0.8, целые: 2 + 1 = 3", 10},
				{"Заполните пропуск: 1/3 = ___/9", db.QuestionTypeFill, []string{}, "3", "Чтобы 1/3 привести к знаменателю 9, нужно и числитель, и знаменатель умножить на 3", 10},
				{"Чему равна разность 3/5 - 1/5?", db.QuestionTypeChoice, []string{"2/5", "1/5", "2/10", "4/5"}, "2/5", "При вычитании дробей с одинаковым знаменателем вычитаем числители: 3-1=2", 10},
				{"Переведите дробь 1/8 в десятичную дробь.", db.QuestionTypeOpen, []string{}, "0.125", "1/8 = 0.125, так как 1 ÷ 8 = 0.125", 15},
				{"Какое число больше: 0.6 или 3/5?", db.QuestionTypeChoice, []string{"Они равны", "0.6", "3/5", "Невозможно сравнить"}, "Они равны", "3/5 = 0.6, это одно и то же число", 15},
				{"Вычислите: 5 - 1.75", db.QuestionTypeOpen, []string{}, "3.25", "5.00 - 1.75 = 3.25", 10},
			},
		},
		{
			ID:          uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			Title:       "Математика 6 класс: Пропорции и отрицательные числа",
			Description: "Тест на знание пропорций, координат и действий с отрицательными числами",
			Subject:     "Математика",
			Grade:       6,
			TimeLimit:   900,
			Questions: []struct {
				Text          string
				Type          db.QuestionType
				Options       []string
				CorrectAnswer string
				Explanation   string
				Points        int
			}{
				{"Решите уравнение: x + 5 = 2", db.QuestionTypeOpen, []string{}, "-3", "x = 2 - 5 = -3", 10},
				{"Чему равно произведение (-3) × (-4)?", db.QuestionTypeChoice, []string{"12", "-12", "7", "-7"}, "12", "При умножении двух отрицательных чисел результат положительный: (-3)×(-4)=12", 10},
				{"Какое число соответствует точке A на координатной прямой? (изображение точки A слева от 0 на расстоянии 3 делений)", db.QuestionTypeFill, []string{}, "-3", "Точка слева от нуля на расстоянии 3 единиц имеет координату -3", 10},
				{"Найдите x из пропорции: 2/x = 4/10", db.QuestionTypeChoice, []string{"5", "4", "2", "8"}, "5", "x = 2 × 10 / 4 = 20/4 = 5", 15},
				{"Чему равна сумма (-7) + 3?", db.QuestionTypeOpen, []string{}, "-4", "-7 + 3 = -4 (движемся вправо по числовой прямой на 3 единицы)", 10},
				{"Вычислите: |-5| + |3|", db.QuestionTypeChoice, []string{"8", "2", "-2", "-8"}, "8", "|-5| = 5, |3| = 3, значит 5 + 3 = 8", 10},
				{"Какой знак нужно поставить: -8 ___ -5 (чтобы получить верное утверждение)", db.QuestionTypeFill, []string{}, "<", "-8 меньше -5, так как -8 лежит левее на числовой прямой", 10},
				{"Решите: (-10) ÷ (-2)", db.QuestionTypeChoice, []string{"5", "-5", "20", "-20"}, "5", "При делении двух отрицательных чисел результат положительный: (-10)÷(-2)=5", 15},
			},
		},
		{
			ID:          uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
			Title:       "Математика 7 класс: Одночлены и многочлены",
			Description: "Тест на знание одночленов, многочленов и формул сокращённого умножения",
			Subject:     "Математика",
			Grade:       7,
			TimeLimit:   900,
			Questions: []struct {
				Text          string
				Type          db.QuestionType
				Options       []string
				CorrectAnswer string
				Explanation   string
				Points        int
			}{
				{"Упростите: 3x² × 2x", db.QuestionTypeOpen, []string{}, "6x³", "3×2=6, x²×x=x³ (при умножении степени складываются)", 10},
				{"Раскройте скобки: (x + 3)²", db.QuestionTypeChoice, []string{"x² + 6x + 9", "x² + 9", "x² + 6x + 6", "x² + 3x + 9"}, "x² + 6x + 9", "По формуле (a+b)² = a² + 2ab + b²: x² + 2·x·3 + 3² = x² + 6x + 9", 15},
				{"Приведите подобные слагаемые: 5a + 3b - 2a + b", db.QuestionTypeOpen, []string{}, "3a + 4b", "5a - 2a = 3a; 3b + b = 4b", 10},
				{"Чему равна степень одночлена 7x³y²?", db.QuestionTypeFill, []string{}, "5", "Степень одночлена = сумма степеней всех переменных: 3 + 2 = 5", 10},
				{"Разложите на множители: x² - 16", db.QuestionTypeChoice, []string{"(x-4)(x+4)", "(x-8)(x+8)", "(x-4)²", "(x+4)²"}, "(x-4)(x+4)", "x² - 16 = x² - 4² = (x-4)(x+4) по формуле разности квадратов", 15},
				{"Выполните умножение: (x - 2)(x + 3)", db.QuestionTypeOpen, []string{}, "x² + x - 6", "(x-2)(x+3) = x·x + x·3 - 2·x - 2·3 = x² + 3x - 2x - 6 = x² + x - 6", 15},
				{"Какое выражение называется одночленом?", db.QuestionTypeChoice, []string{"Произведение чисел и переменных", "Сумма одночленов", "Выражение с делением на переменную", "Уравнение"}, "Произведение чисел и переменных", "Одночлен — это произведение чисел, переменных и их степеней", 10},
				{"Упростите: (2x)³", db.QuestionTypeFill, []string{}, "8x³", "(2x)³ = 2³ · x³ = 8x³", 15},
			},
		},
	}

	for _, q := range quizzes {
		quiz := db.CreateQuizParams{
			ID:          q.ID,
			Title:       q.Title,
			Description: q.Description,
			Subject:     q.Subject,
			Grade:       q.Grade,
			Status:      db.QuizStatusAvailable,
			TimeLimit:   q.TimeLimit,
			CreatedBy:   teacherID,
			CreatedAt:   time.Now(),
		}
		if _, err := queries.CreateQuiz(ctx, quiz); err != nil {
			return fmt.Errorf("failed to create quiz: %w", err)
		}

		for i, qq := range q.Questions {
			optionsJSON, _ := json.Marshal(qq.Options)
			question := db.CreateQuestionParams{
				ID:            uuid.New(),
				QuizID:        q.ID,
				Text:          qq.Text,
				Type:          qq.Type,
				Options:       optionsJSON,
				CorrectAnswer: qq.CorrectAnswer,
				Explanation:   qq.Explanation,
				Points:        qq.Points,
				OrderIndex:    i,
			}
			if _, err := queries.CreateQuestion(ctx, question); err != nil {
				return fmt.Errorf("failed to create question: %w", err)
			}
		}
	}

	log.Println("Seeding completed")
	return nil
}

type jsonQuizInput struct {
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	Subject        string             `json:"subject"`
	Grade          int                `json:"grade"`
	TimeLimit      int                `json:"time_limit"`
	CreatedByEmail string             `json:"created_by_email"`
	Questions      []jsonQuestionInput `json:"questions"`
}

type jsonQuestionInput struct {
	Text          string   `json:"text"`
	Type          string   `json:"type"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correct_answer"`
	Explanation   string   `json:"explanation"`
	Points        int      `json:"points"`
	OrderIndex    int      `json:"order_index"`
}

func LoadInitialDataFromFolder(ctx context.Context, pool *pgxpool.Pool, folder string) error {
	queries := db.New(pool)

	entries, err := os.ReadDir(folder)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read initial data folder: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(folder, entry.Name()))
		if err != nil {
			log.Printf("Warning: failed to read %s: %v", entry.Name(), err)
			continue
		}

		var quizzes []jsonQuizInput
		if err := json.Unmarshal(data, &quizzes); err != nil {
			log.Printf("Warning: failed to parse %s: %v", entry.Name(), err)
			continue
		}

		for _, q := range quizzes {
			existingQuizzes, err := queries.GetQuizzes(ctx)
			if err != nil {
				log.Printf("Warning: failed to get quizzes: %v", err)
				continue
			}
			quizExists := false
			for _, existing := range existingQuizzes {
				if existing.Title == q.Title {
					quizExists = true
					break
				}
			}
			if quizExists {
				log.Printf("Quiz '%s' already exists, skipping", q.Title)
				continue
			}

			teacher, err := queries.GetUserByEmail(ctx, q.CreatedByEmail)
			if err != nil {
				log.Printf("Warning: teacher with email %s not found, skipping quiz '%s'", q.CreatedByEmail, q.Title)
				continue
			}

			quiz := db.CreateQuizParams{
				ID:          uuid.New(),
				Title:       q.Title,
				Description: q.Description,
				Subject:     q.Subject,
				Grade:       q.Grade,
				Status:      db.QuizStatusAvailable,
				TimeLimit:   q.TimeLimit,
				CreatedBy:   teacher.ID,
				CreatedAt:   time.Now(),
			}
			if _, err := queries.CreateQuiz(ctx, quiz); err != nil {
				log.Printf("Warning: failed to create quiz '%s': %v", q.Title, err)
				continue
			}

			for _, qq := range q.Questions {
				optionsJSON, _ := json.Marshal(qq.Options)
				question := db.CreateQuestionParams{
					ID:            uuid.New(),
					QuizID:        quiz.ID,
					Text:          qq.Text,
					Type:          db.QuestionType(qq.Type),
					Options:       optionsJSON,
					CorrectAnswer: qq.CorrectAnswer,
					Explanation:   qq.Explanation,
					Points:        qq.Points,
					OrderIndex:    qq.OrderIndex,
				}
				if _, err := queries.CreateQuestion(ctx, question); err != nil {
					log.Printf("Warning: failed to create question in quiz '%s': %v", q.Title, err)
				}
			}
			log.Printf("Loaded quiz '%s' from %s", q.Title, entry.Name())
		}
	}

	return nil
}
