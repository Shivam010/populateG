# populateG | A Web Application To Populate Google Docs Template 

Use the application to populate the your own Google Docs Template by using the data present in your Google Spreadsheets.
Just tag your Document using the headers used in your Google Spreadsheets.

See the example of the Template Document and the Data Spreadsheet: [**here**](https://drive.google.com/open?id=1o5p9L2EOf6WPP_DiZxvo380E7nl7SO71)

[**Try it here**](https://populateg.herokuapp.com/)

## Sample Example
![How to use populateG](images/populateg.gif)

## How to tag the template?
As shown in above gif, to create template, you have to tag your document with the headers specified in your spreadsheet by enclosing them into 2 curly braces, **`{{`**.

Like, if the header is **`name`**, then use **`{{name}}`** as the tag.

## Run Locally
Just, setup GoLang in your system.<br>
And Run `go get github.com/Shivam010/populateG` or just clone the repo. Now, run `go run .` in the project.<br>
If in $GOPATH use `GO111MODULE=on go run .`

## Request for Contribution
Changes and improvements are more than welcome! 

Feel free to fork and pull a request [here.](https://github.com/Shivam010/populateG)

Please make your changes in a specific branch and request to pull into master! If you can please make sure that the changes work properly and does not affect the functioning of the website before sending a Pull Request, as that will help speed up the process.

## License
The application, its design and its code all are licensed under the [MIT license.](https://github.com/Shivam010/populateG/blob/master/LICENSE) 
