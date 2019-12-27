import React, { useState, useCallback, useEffect } from "react";
import FileManager, { Permissions } from 'devextreme-react/file-manager';
import { FileUploader } from 'devextreme-react';
import API from '../../api';
import FileViewer from 'react-file-viewer';


const HomePage = props => {
    const [open, setOpen] = React.useState(false);
    const [currentPath, setCurrentPath] = React.useState("documents");
    const [fileSystem, setFileSystem] = React.useState();
    //const [documentToDisplay, setDocumentToDisplay] = React.useState("") //URL to document to display

    //useEffect(async () => {
    //    const fileSystem = await API.get(`/documents/filesystem`);
    //})
    var fixAPI
    useEffect(() => {
        // Update the document title using the browser API
        API.get(`/documents/filesystem`).then(result => checkData(result.data))
      }, []);  
    const allowedFileExtensions = ['.doc', '.docx', '.pdf', '.txt', '.odf', '.rtf', '.png', '.jpeg', '.tiff', '.jpg']

    const checkData = (data) => {
        fixAPI = [data]
        setFileSystem(fixAPI)
    }

    function openInNewTab(url) {
        var win = window.open(url, '_blank');
        win.focus();
      }

    const openFile = event => {
        var baseURL
        var url
        console.log(event)
        console.log(event.fileItem.dataItem.documentURL)
        baseURL = API.defaults.baseURL
        url = baseURL + event.fileItem.dataItem.documentURL
        console.log(url)
        openInNewTab(url)

        
    }

    const directoryChanged = event => {
        setCurrentPath(event.component.option('currentPath'))
        console.log("Path", event.component.option('currentPath'))
    }
    const checkUpload = event => {
        console.log("Upload Event", event)
    }


    const toolbarCustomization = {items: [{name: "showNavPane", visible: true}, "separator", "create", {name: "upload", visible: true }, {name: "create", visible: true}, "refresh", {name: "separator", location: "after"}, "viewSwitcher"]}

    return (
        <React.Fragment>
            <FileUploader
            visible={false}
            ></FileUploader>
            <FileManager
                onOptionChanged={checkUpload}
                fileProvider={fileSystem}
                onCurrentDirectoryChanged={directoryChanged}
                allowedFileExtensions={allowedFileExtensions}
                onSelectedFileOpened={openFile}
                toolbar={toolbarCustomization}
                currentPath={currentPath}
                >
                <Permissions
                    create={false}
                    copy={true}
                    move={true}
                    remove={true}
                    rename={true}
                    upload={true}
                    download={true}>
                </Permissions>             
            </FileManager>
        </React.Fragment>

    );
  };
  
export default HomePage;