import React, { useState, useEffect } from "react";
import 'chonky/style/main.css';
import {FileBrowser, FileView} from 'chonky';
import { makeStyles } from '@material-ui/core/styles';
import API from '../../api';
import FileViewer from 'react-file-viewer';
import { Button, TextField, Modal, Typography } from '@material-ui/core';
import 'react-dropzone-uploader/dist/styles.css'
import Dropzone from 'react-dropzone-uploader'
import { toast } from 'react-toastify';

const useStyles = makeStyles(theme => ({
    viewer: {
      position: 'absolute',
      left: '20%',
      marginTop: '5%',
      width: '50%',
      height: '80%',
      backgroundColor: theme.palette.background.paper,
      border: '2px solid #000',
      boxShadow: theme.shadows[5],
      padding: theme.spacing(2, 4, 3),
    },
    paper: {
        position: 'absolute',
        left: '1%',
        marginTop: '1%',
        backgroundColor: theme.palette.background.paper,
        border: '2px solid #000',
        boxShadow: theme.shadows[5],
        padding: theme.spacing(2, 4, 3),
    },
    buttons: {
        '& > *': {
            margin: theme.spacing(1),
        }
    }
  }));


const HomePage = props => {
    const classes = useStyles();
    const [currentfileTree, setCurrentFileSystem] = React.useState([]);
    const [fullFileSystem, setFullFileSystem] = React.useState([]);
    const [folderTree, setFolderSystem] = React.useState([]);
    const [currentFolder, setCurrentFolder] = React.useState("");
    const [modalOpen, setModalOpen] = React.useState(false)
    const [uploadOpen, setUploadOpen] = React.useState(false)
    const [folderModalOpen, setfolderModalOpen] = React.useState(false)
    const [folderName, setFolderName] = React.useState("Empty")
    const [currentFileExt, setCurrentFileExt] = React.useState("")
    const [currentFileURL, setCurrentFileURL] = React.useState("")
    const [currentSelection, setCurrentSelection] = React.useState([])
    const [deleteModalOpen, setDeleteFolderModal] = React.useState(false)
    

    //const allowedFileTypes = ['application/msword', '.application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'application/pdf', 'text/plain', 'application/vnd.oasis.opendocument.text', 'application/rtf', 'image/*']
    const allowedFileTypes = "application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/pdf,text/plain,application/vnd.oasis.opendocument.text,application/rtf,image/*"
    
 
    // specify upload params and url for your files
    const getUploadParams = ({ file }) => { 
        const body = new FormData()
        let filePath = ""
        folderTree.shift() //skip the root folder since that is the one that is defined in the backend
       folderTree.forEach(folder => {
            filePath = filePath + folder.name + "/"
        })
        body.append('file', file)
        body.append('path', filePath)
        return { url: API.defaults.baseURL + '/document/upload', body } 
    }

    // called every time a file's `status` changes //TODO remove, this is for testing only
    const handleChangeStatus = ({ meta, file }, status) => { console.log(status, meta, file) }
    
    // receives array of files that are done uploading when submit button is clicked
    const handleSubmit = (files, allFiles) => {
        console.log(files.map(f => f.meta))
        //allFiles.forEach(API.put('document/upload'))
        allFiles.forEach(f => f.remove())
    }

    useEffect(() => { //On page load
        getFileSystem()
      }, []); 

    useEffect(() => {  //when a search occurs
        console.log("Search Terms2: ", props.searchTerm)
        if (props.searchTerm.length > 3) {
        API.get(`/search/?term=` + props.searchTerm).then(result => checkData(result.data))
        }
        if (props.searchTerm.length == 0) {
        API.get(`/documents/filesystem`).then(result => checkData(result.data))
        }
    }, [props.searchTerm]);


    const getFileSystem = () => {
        API.get(`/documents/filesystem`).then(result => checkData(result.data))
    }

    const checkData = (data) => {
        if (data[0] == null) {
            console.log("Undefined Data! ", data[0])
            toast.warn("No Results found! " + props.searchTerm);
            return
        }
        if (props.searchTerm.length > 0) {
            toast.success("Results Found! " + props.searchTerm);
        }
        console.log("HERE IS DATA: ", data)
        setCurrentFolder(data[0].id) //folder we are in
        setFolderSystem([data[0]]) //folder trail at top
        setFullFileSystem(data) 
        console.log("Full file system in check data: ", fullFileSystem)
        getFiles(data[0], data)
    }

      const getFiles = (folder, files) => {
          console.log("FolderID", folder.id)
          console.log("Files", files)
          //childIDs = folder.childrenIDs
          var file
          const fileResults = files.filter(function(file) {
            if (file.parentID == folder.id) {
                return file
            }
          })
          console.log("FileResults ", fileResults)
          setCurrentFileSystem(fileResults)
      }

    function openInNewTab(url) { //TODO perhaps context menu to open in new tab?
        var win = window.open(url, '_blank');
        win.focus();
      }

    const checkUpload = event => {
        console.log("Upload Event", event)
    }
    const onError = (event) => {
        console.log("Error in viewer", event)
    }

    const fileViewer = (file) => {
        let fileExt = file.name.split('.').pop()
        let fileURL = API.defaults.baseURL + file.fileURL
        setCurrentFileExt(fileExt)
        setCurrentFileURL(fileURL)
        setModalOpen(true)
            
    }

    const handleFolderCreate = () => {
        setfolderModalOpen(false)
        console.log("FolderName", folderName)
        console.log("FolderTree", folderTree)
        let filePath = ""
        folderTree.shift() //skip the root folder
       folderTree.forEach(folder => {
            filePath = filePath + folder.name + "/"
        })
        console.log("FILEPATH", filePath)
        API.post('/folder/?path=' + filePath + '&folder=' + folderName).then(postResult => { //TODO give response to folder create
            API.get(`/documents/filesystem`).then(result => {checkData(result.data)}) //TODO reset path to the new folder? Will be hard to do
        })      
    }

    const handleDownload = () => {
        console.log("Selection", typeof(currentSelection))
        let fileArray = []
        Object.keys(currentSelection).forEach(value => {
            fullFileSystem.filter(file => {
                if (file.id == value) {
                    fileArray.push(file)
                }
            })
        })
        console.log("File Array", fileArray)
        fileArray.forEach(file => {
            let href = API.defaults.baseURL + file.fileURL;
            fetch(href).then(response => {
                response.blob().then(blob => {
                    let url = window.URL.createObjectURL(blob);
                    let a = document.createElement('a');
                    a.href = url;
                    a.download = file.name;
                    a.click();
                })
            })
        })
    }

    const handleDeleteFiles = (files) => {
        let filePath = ""
        files.forEach(file => {
            if (folderTree.length == 1) { //we are at root folder
                filePath = file.name
            } else {
                folderTree.shift() //skip the root folder since that is the one that is defined in the backend
                folderTree.forEach(folder => {
                    filePath = filePath + folder.name + "/"
                })
                filePath = filePath + file.name
            }
            console.log("FilePath", filePath)
            API.delete('/document/?path=' + filePath + '&id=' + file.id)
            filePath = ""
        })
        API.get(`/documents/filesystem`).then(result => {checkData(result.data)}) //TODO reset path to the new folder? Will be hard to do
    }

    const handleFileOpen = (file) => {
        if (file.isDir) {
            console.log("SETTING DIR")
            let tempFile = file
            let folderTreeNew = [file] //add the current folder to the array
            let inFolder = file
            while (inFolder.parentID != "") {
                fullFileSystem.find(function(currentFile, index) {
                    if (currentFile.id == tempFile.parentID) {
                        inFolder = currentFile
                        folderTreeNew.unshift(currentFile)
                        tempFile = currentFile
                    }
                })
            }
            setFolderSystem(folderTreeNew)
            getFiles(file, fullFileSystem)
        } else { // if it is a file not a folder
            console.log("working on file", file.name)
            fileViewer(file)
        }
    }
    
    return (
        <React.Fragment>
        <FileBrowser
            files={currentfileTree}
            folderChain={folderTree}
            onFileOpen={handleFileOpen}
            view={FileView.SmallThumbs}
            onFolderCreate={() => setfolderModalOpen(true)}
            onDeleteFiles={() => setDeleteFolderModal(true)}
            onUploadClick={() => setUploadOpen(true)}
            onDownloadFiles={handleDownload}
            onSelectionChange={(selection) => setCurrentSelection(selection)}
            />
        <Modal open={modalOpen} onClose={() => setModalOpen(false)}>
            <div className={classes.viewer}>
                <FileViewer
                        fileType={currentFileExt}
                        filePath={currentFileURL}
                        onError={onError}/>
            </div>
        </Modal>
        <Modal open={uploadOpen} onClose={() => setUploadOpen(false)}>
            <div className={classes.viewer}>
                <Dropzone
                    getUploadParams={getUploadParams}
                    onChangeStatus={handleChangeStatus}
                    onSubmit={handleSubmit}
                    accept={allowedFileTypes}
                    />
            </div>
        </Modal>
        <Modal open={folderModalOpen} onClose={() => setfolderModalOpen(false)}>
            <div className={classes.paper}>
                <TextField id="folderName" label="Folder Name" onChange={(event) => setFolderName(event.target.value)} />
                <Button variant="contained" onClick={handleFolderCreate}>Submit</Button>
            </div>
        </Modal>
        <Modal open={deleteModalOpen} onClose={() => setDeleteFolderModal(false)}>
            <div className={classes.paper}>
                <Typography>Are you sure you want to delete these files?</Typography>
                <div className={classes.buttons}>
                    <Button variant="contained" onClick={() => setDeleteFolderModal(false)}>Cancel</Button>
                    <Button variant="contained" color="secondary" onClick={handleDeleteFiles}>Delete</Button>
                </div>
            </div>
        </Modal>
        </React.Fragment>
    )  
}

  
export default HomePage;