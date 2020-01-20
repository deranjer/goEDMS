import axios from 'axios';

const { apiUrl } = window['runConfig'];
console.log("API URL: ", apiUrl)
export default axios.create({
    baseURL: apiUrl
})