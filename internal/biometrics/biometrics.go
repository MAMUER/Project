package biometrics

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

type BiometricData struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    DeviceType string   `json:"device_type"`
    HeartRate int       `json:"heart_rate"`
    ECG       string    `json:"ecg"`
    Systolic  int       `json:"systolic"`
    Diastolic int       `json:"diastolic"`
    SpO2      int       `json:"spo2"`
    Temperature float64 `json:"temperature"`
    SleepDuration int   `json:"sleep_duration"`
    DeepSleep   int     `json:"deep_sleep"`
    Timestamp   time.Time `json:"timestamp"`
    CreatedAt   time.Time `json:"created_at"`
}

type Processor struct {
    db     *sql.DB
    mq     *amqp.Channel
    queue  string
}

func NewProcessor(db *sql.DB, mq *amqp.Channel, queue string) *Processor {
    return &Processor{
        db:     db,
        mq:     mq,
        queue:  queue,
    }
}

func (p *Processor) Save(ctx context.Context, data *BiometricData) error {
    query := `
        INSERT INTO biometric_data 
        (id, user_id, device_type, heart_rate, ecg, systolic, diastolic, spo2, temperature, sleep_duration, deep_sleep, timestamp)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `
    _, err := p.db.ExecContext(ctx, query,
        data.ID, data.UserID, data.DeviceType, data.HeartRate, data.ECG,
        data.Systolic, data.Diastolic, data.SpO2, data.Temperature,
        data.SleepDuration, data.DeepSleep, data.Timestamp,
    )
    return err
}

func (p *Processor) Publish(ctx context.Context, data *BiometricData) error {
    body, err := json.Marshal(data)
    if err != nil {
        return err
    }

    return p.mq.PublishWithContext(ctx,
        "", p.queue, false, false,
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
            Timestamp:    time.Now(),
        })
}