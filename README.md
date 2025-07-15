ポートフォリオ用に公開しました、大学の学士論文のために作ったCLIプログラムです。

このプログラムはクリーンコードをサポートするためにAIを使ってJavaのソースコードで関数名を入力した場合、戻り値や引数をユーザーに提案します。Visual Studio Codeなどの拡張機能として作られています。プロジェクト自体は以下の三つのソフトで構成されています。

- [Java Crawler](https://github.com/jkahlem/javacrawler): Javaのソースコードを解析して、必要なデータや構成要素を抽出しXMLに変換するプログラムです
- **Cleancode Util**: Goで書かれたソフトで、以下の二つのソフトに分けられています。
  - **Data Extractor**: トレーニングセットや検証セットを作るCLIです。指定されたレポジトリをクローンして、Java Crawlerを使ってデータを抽出・解析し、決まったルールでデータセットを作成します。それらのセットをPredictorに送信することでAIを学習させられます。
  - **Language Server**: Language Server Protocol (LSP)を実装したサーバーアプリで、リアルタイムでプロジェクトの解析をし、入力される関数名をPredictorに送信し、推測された戻り値や引数を提案としてVisual Studio Codeに送ります。
- [Predictor](https://github.com/jkahlem/predictor): AIモデルの操作（学習・推測）のためのプログラムです。

## 主なフロー

### 学習
https://github.com/jkahlem/cleancodeutil/blob/main/processing/Processing.go
1. DataExtractorで指定のレポジトリをクローンする
2. DataExtractorでデータをJavaCrawlerに送り、XMLに変換する
3. データの準備 (ソースコードで使われたタイプを正規名に直し、関数の戻り値・引数などをタイプとリンクさせる)
   https://github.com/jkahlem/cleancodeutil/blob/main/common/code/java/Resolver.go
4. データセット（フィルター設定などによって一部の関数名などは除外される）やデータセットに関するデータ(RougeやBleuなどの評価指標)を作成する    https://github.com/jkahlem/cleancodeutil/blob/main/processing/dataset/TrainingDatasetCreator.go
5. データセットをPredictorに送信して、AIモデルを学習させる

### 拡張機能
https://github.com/jkahlem/cleancodeutil/blob/main/languageserver/LanguageServer.go
https://github.com/jkahlem/cleancodeutil/blob/main/languageserver/Controller.go
1. Language ServerとVisual Studio Codeの通信の確立、プロジェクト全体を解析する
2. Visual Studio Codeでの入力に応じて、入力された関数名をPredictorに送信して、戻り値や引数（名前とタイプ）を生成する
3. Diagnosticsを作成しVisual Studio Codeに送ることで戻り値や引数の自動入力の提案を表示させる

実行のために設定ファイルによって一部の設定が必要です。設定の詳細はこちらに書いてあります。
https://github.com/jkahlem/cleancodeutil/blob/main/common/configuration/ConfigFile.go
