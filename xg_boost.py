import pandas as pd
import xgboost as xgb
from sklearn.model_selection import train_test_split
import joblib
import matplotlib.pyplot as plt
from graphviz import Source
import os

# Prepare the data
try:
    df = pd.read_csv("dataset.csv")
except FileNotFoundError:
    print("Error: dataset.csv not found in current directory")
    exit(1)

df = df.fillna(0)
df = df.drop(columns=["id", "team_1_id", "team_2_id", "date"])
df = df.iloc[:1000]

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


# Define features and target
X = cleaned_df.drop(columns=["team_won_label"])
y = cleaned_df["team_won_label"]

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

# Initialize and train XGBoost Classifier
xgb_model = xgb.XGBClassifier(
    n_estimators=100,
    max_depth=5,
    learning_rate=0.1,
    subsample=0.8,
    colsample_bytree=0.8,
    random_state=42,
    eval_metric="logloss",
)

# Train the model
xgb_model.fit(X, y)

# Export the first 5 decision trees as PNG images
num_trees_to_export = 5
for tree_idx in range(min(num_trees_to_export, xgb_model.n_estimators)):
    try:
        plt.figure(figsize=(20, 10))  # Set figure size
        xgb.plot_tree(xgb_model, tree_idx=tree_idx, rankdir="LR")  # Use tree_idx
        output_file = f"tree_{tree_idx}.png"
        plt.savefig(output_file, format="png", bbox_inches="tight", dpi=300)
        plt.close()
        print(f"Saved tree {tree_idx} as {output_file}")
    except Exception as e:
        print(f"Error exporting tree {tree_idx}: {str(e)}")
        plt.close()
        continue

# Save the model
joblib.dump(xgb_model, "xgboost_model.pkl")
print("\nModel saved as 'xgboost_model.pkl'")
