import pandas as pd
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import cross_val_score, cross_validate, train_test_split
from sklearn.preprocessing import StandardScaler
from sklearn.metrics import classification_report, confusion_matrix
import joblib
import numpy as np

# Prepare the data
try:
    df = pd.read_csv("dataset.csv")
except FileNotFoundError:
    print("Error: dataset.csv not found in /home/leminhohoho/repos/sport-prediction/")
    exit(1)

df = df.fillna(0)
df = df.drop(columns=["id", "team_1_id", "team_2_id", "date"])

# Drop low-importance features (from Random Forest output)
low_importance_features = [
    "team_1_win_rate_on_Breeze",
    "team_2_win_rate_on_Breeze",
    "team_1_played_times_on_Breeze",
    "team_2_played_times_on_Breeze",
    "team_1_win_rate_on_Sunset",
]
df = df.drop(columns=low_importance_features, errors="ignore")

# Define features and target
X = df.drop(columns=["team_won_label"])
y = df["team_won_label"]

# Map labels from [1, 2] to [0, 1]
y = y.map({1: 0, 2: 1})

# Verify label mapping
print("Unique labels after mapping:", y.unique())
if not set(y.unique()).issubset({0, 1}):
    raise ValueError(
        "Label mapping failed! Expected labels [0, 1], got {}".format(y.unique())
    )

# Check class distribution
print("Class Distribution:")
print(y.value_counts(normalize=True))

# Scale features
scaler = StandardScaler()
X_scaled = scaler.fit_transform(X)
X_scaled = pd.DataFrame(X_scaled, columns=X.columns)  # Retain column names

# Initialize Logistic Regression
log_reg = LogisticRegression(
    max_iter=1000,  # Increased iterations for convergence
    class_weight="balanced",  # Handle slight class imbalance
    random_state=42,
    solver="lbfgs",  # Default solver, good for small datasets
)

# Perform cross-validation
cv_scores = cross_val_score(
    log_reg, X_scaled, y, cv=5, scoring="accuracy", error_score="raise"
)
print("Cross-Validation Accuracy Scores:", cv_scores)
print("Mean CV Accuracy:", np.mean(cv_scores))
print("Standard Deviation:", np.std(cv_scores))

# Detailed cross-validation
scoring = {
    "accuracy": "accuracy",
    "precision": "precision_weighted",
    "recall": "recall_weighted",
    "f1": "f1_weighted",
}
cv_results = cross_validate(log_reg, X_scaled, y, cv=5, scoring=scoring)

# Print detailed statistics
print("\nDetailed Cross-Validation Results:")
for metric in cv_results:
    print(f"{metric}:")
    print(f"  Scores: {cv_results[metric]}")
    print(f"  Mean: {np.mean(cv_results[metric]):.3f}")
    print(f"  Std: {np.std(cv_results[metric]):.3f}")

# Train-test split for confusion matrix and detailed metrics
X_train, X_test, y_train, y_test = train_test_split(
    X_scaled, y, test_size=0.2, random_state=42
)
log_reg.fit(X_train, y_train)
y_pred = log_reg.predict(X_test)

# Confusion matrix
print("\nConfusion Matrix (rows: actual, columns: predicted):")
print(confusion_matrix(y_test, y_pred, labels=[0, 1]))

# Classification report
print("\nClassification Report:")
print(
    classification_report(
        y_test, y_pred, target_names=["Team 1 Win (0)", "Team 2 Win (1)"]
    )
)

# Feature coefficients
feature_coefficients = pd.DataFrame(
    {"feature": X.columns, "coefficient": log_reg.coef_[0]}
).sort_values("coefficient", key=abs, ascending=False)
print("\nFeature Coefficients:")
print(feature_coefficients)

# Save the model and scaler
joblib.dump(log_reg, "logistic_regression_model.pkl")
joblib.dump(scaler, "scaler.pkl")
print("\nModel saved as 'logistic_regression_model.pkl'")
print("Scaler saved as 'scaler.pkl'")
