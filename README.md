# gha2gss

## 概要
- GitHub Actionの実行結果ログをGitHub APIから取得し、GoogleSpreadSheetに保存する

## ユースケース
- 変更障害率の計算などに用いることを想定

## 環境変数
- `$GITHUB_TOKEN`
  - personal access token
- `$GITHUB_ACTION_CONFIG_URLS`
  - comma separated values
  - ex. https://github.com/foo/bar/blob/master/.github/workflows/cd.yml,https://github.com/xxxx/yyyy/blob/master/.github/workflows/cd.yml
- `$GOOGLE_APPLICATION_CREDENTIALS`
  - filepath of google application credentials
- `SPREADSHEET_ID`
  - id of google spheadsheet
- `$TARGET_SHEET_NAME`
