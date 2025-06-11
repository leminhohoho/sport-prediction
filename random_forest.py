import pandas as pd
import pydotplus
from sklearn.model_selection import train_test_split
from sklearn.tree import export_graphviz
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score, classification_report
from graphviz import Source
from sklearn.model_selection import StratifiedKFold, GridSearchCV
from sklearn.model_selection import cross_val_score, cross_validate
from sklearn.metrics import classification_report, confusion_matrix
import joblib
import numpy as np
import os

# Prepare the data
df = pd.read_csv("dataset2.csv")
df = df.fillna(0)
df = df.drop(columns=["id", "team_1_id", "team_2_id", "date"])
df = df.iloc[:500]

cleaned_df = pd.DataFrame(index=df.index)

cleaned_df["win_rate_diff"] = df["team_1_win_rate"] - df["team_2_win_rate"]
cleaned_df["match_played_diff"] = df["team_1_match_played"] - df["team_2_match_played"]
cleaned_df["avg_rating_diff"] = (
    df["team_1_player_1_avg_rating"]
    + df["team_1_player_2_avg_rating"]
    + df["team_1_player_3_avg_rating"]
    + df["team_1_player_4_avg_rating"]
    + df["team_1_player_5_avg_rating"]
    - df["team_2_player_1_avg_rating"]
    - df["team_2_player_2_avg_rating"]
    - df["team_2_player_3_avg_rating"]
    - df["team_2_player_4_avg_rating"]
    - df["team_2_player_5_avg_rating"]
) / 5
# NOTE: Compare the avg rating between 2 player from 2 team that has the highest rating
cleaned_df["highest_avg_rating_diff"] = df[
    [
        "team_1_player_1_avg_rating",
        "team_1_player_2_avg_rating",
        "team_1_player_3_avg_rating",
        "team_1_player_4_avg_rating",
        "team_1_player_5_avg_rating",
    ]
].max(axis=1) - df[
    [
        "team_2_player_1_avg_rating",
        "team_2_player_2_avg_rating",
        "team_2_player_3_avg_rating",
        "team_2_player_4_avg_rating",
        "team_2_player_5_avg_rating",
    ]
].max(
    axis=1
)

# NOTE: Compare the avg acs between 2 player from 2 team that has the highest acs
cleaned_df["highest_avg_acs_diff"] = df[
    [
        "team_1_player_1_avg_acs",
        "team_1_player_2_avg_acs",
        "team_1_player_3_avg_acs",
        "team_1_player_4_avg_acs",
        "team_1_player_5_avg_acs",
    ]
].max(axis=1) - df[
    [
        "team_2_player_1_avg_acs",
        "team_2_player_2_avg_acs",
        "team_2_player_3_avg_acs",
        "team_2_player_4_avg_acs",
        "team_2_player_5_avg_acs",
    ]
].max(
    axis=1
)

# NOTE: Compare the avg kast between 2 teams
cleaned_df["avg_kast_diff"] = (
    df["team_1_player_1_avg_kast"]
    + df["team_1_player_2_avg_kast"]
    + df["team_1_player_3_avg_kast"]
    + df["team_1_player_4_avg_kast"]
    + df["team_1_player_5_avg_kast"]
    + df["team_2_player_1_avg_kast"]
    + df["team_2_player_2_avg_kast"]
    + df["team_2_player_3_avg_kast"]
    + df["team_2_player_4_avg_kast"]
    + df["team_2_player_5_avg_kast"]
) / 5

# NOTE: Compare the avg adr between 2 player from 2 team that has the highest adr
cleaned_df["highest_avg_adr_diff"] = df[
    [
        "team_1_player_1_avg_adr",
        "team_1_player_2_avg_adr",
        "team_1_player_3_avg_adr",
        "team_1_player_4_avg_adr",
        "team_1_player_5_avg_adr",
    ]
].max(axis=1) - df[
    [
        "team_2_player_1_avg_adr",
        "team_2_player_2_avg_adr",
        "team_2_player_3_avg_adr",
        "team_2_player_4_avg_adr",
        "team_2_player_5_avg_adr",
    ]
].max(
    axis=1
)

# NOTE: Compare the avg first kills between 2 players from 2 teams those have the highest first kills
cleaned_df["highest_avg_first_kills_diff"] = df[
    [
        "team_1_player_1_avg_first_kills",
        "team_1_player_2_avg_first_kills",
        "team_1_player_3_avg_first_kills",
        "team_1_player_4_avg_first_kills",
        "team_1_player_5_avg_first_kills",
    ]
].max(axis=1) - df[
    [
        "team_2_player_1_avg_first_kills",
        "team_2_player_2_avg_first_kills",
        "team_2_player_3_avg_first_kills",
        "team_2_player_4_avg_first_kills",
        "team_2_player_5_avg_first_kills",
    ]
].max(
    axis=1
)

# NOTE: Compare the avg first deaths between 2 players from 2 teams those have the highest first deaths
cleaned_df["highest_avg_first_deaths_diff"] = df[
    [
        "team_1_player_1_avg_first_deaths",
        "team_1_player_2_avg_first_deaths",
        "team_1_player_3_avg_first_deaths",
        "team_1_player_4_avg_first_deaths",
        "team_1_player_5_avg_first_deaths",
    ]
].max(axis=1) - df[
    [
        "team_2_player_1_avg_first_deaths",
        "team_2_player_2_avg_first_deaths",
        "team_2_player_3_avg_first_deaths",
        "team_2_player_4_avg_first_deaths",
        "team_2_player_5_avg_first_deaths",
    ]
].max(
    axis=1
)

# NOTE: Compare the avg multikills between 2 teams
cleaned_df["avg_multikills_diff"] = (
    df["team_1_player_1_avg_multikills"]
    + df["team_1_player_2_avg_multikills"]
    + df["team_1_player_3_avg_multikills"]
    + df["team_1_player_4_avg_multikills"]
    + df["team_1_player_5_avg_multikills"]
    + df["team_2_player_1_avg_multikills"]
    + df["team_2_player_2_avg_multikills"]
    + df["team_2_player_3_avg_multikills"]
    + df["team_2_player_4_avg_multikills"]
    + df["team_2_player_5_avg_multikills"]
) / 5

# NOTE: Compare the avg multikills between 2 players from 2 teams those have the highest multikills
cleaned_df["highest_avg_multikills_diff"] = df[
    [
        "team_1_player_1_avg_multikills",
        "team_1_player_2_avg_multikills",
        "team_1_player_3_avg_multikills",
        "team_1_player_4_avg_multikills",
        "team_1_player_5_avg_multikills",
    ]
].max(axis=1) - df[
    [
        "team_2_player_1_avg_multikills",
        "team_2_player_2_avg_multikills",
        "team_2_player_3_avg_multikills",
        "team_2_player_4_avg_multikills",
        "team_2_player_5_avg_multikills",
    ]
].max(
    axis=1
)
# NOTE: Compare the avg clutches between 2 players from 2 teams those have the highest clutches
cleaned_df["highest_avg_clutches_diff"] = df[
    [
        "team_1_player_1_avg_clutches",
        "team_1_player_2_avg_clutches",
        "team_1_player_3_avg_clutches",
        "team_1_player_4_avg_clutches",
        "team_1_player_5_avg_clutches",
    ]
].max(axis=1) - df[
    [
        "team_2_player_1_avg_clutches",
        "team_2_player_2_avg_clutches",
        "team_2_player_3_avg_clutches",
        "team_2_player_4_avg_clutches",
        "team_2_player_5_avg_clutches",
    ]
].max(
    axis=1
)

cleaned_df["team_won_label"] = df["team_won_label"]

print(cleaned_df)

# Load your dataset (replace with your data loading code)
# cleaned_df = pd.read_csv('your_dataset.csv')
X = cleaned_df.drop(columns=["team_won_label"])  # Features
y = cleaned_df["team_won_label"]  # Target

X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.2, random_state=42
)

rf = RandomForestClassifier(
    n_estimators=100,  # Number of trees
    max_depth=5,  # Maximum depth of each tree
    min_samples_split=2,  # Minimum samples to split a node
    random_state=42,
)
rf.fit(X_train, y_train)

# Make predictions
y_pred = rf.predict(X_test)

# Evaluation metrics
accuracy = accuracy_score(y_test, y_pred)
print(f"\nAccuracy: {accuracy:.2f}")

# Feature importance
feature_importance = pd.DataFrame(
    {"Feature": X.columns, "Importance": rf.feature_importances_}
).sort_values(by="Importance", ascending=False)
print("\nFeature Importance:")
print(feature_importance)

# Save the model
joblib.dump(rf, "random_forest_model.pkl")
print("\nModel saved as 'random_forest_model.pkl'")

# Export all decision trees as images
os.makedirs("decision_trees", exist_ok=True)  # Create directory for trees
for i, tree in enumerate(rf.estimators_):
    dot_data = export_graphviz(
        tree,
        out_file=None,
        feature_names=X.columns,
        class_names=["Team 1", "Team 2"],
        filled=True,
        rounded=True,
        special_characters=True,
    )
    graph = Source(dot_data)
    graph.render(f"decision_trees/tree_{i+1}", format="png", cleanup=True)
print(
    f"\nAll {len(rf.estimators_)} decision trees saved as PNG images in 'decision_trees' directory"
)
