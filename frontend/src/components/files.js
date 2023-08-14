import React, { useState, useEffect } from 'react';
import './styles.css';

const FileForm = () => {
  const [file, setFile] = useState(null);
  const [fileList, setFileList] = useState([]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const formData = new FormData();
      formData.append('file', file);

      const response = await fetch(`/api/upload`, {
        method: 'POST',
        body: formData,
      });

      if (response.ok) {
        window.location.reload();
        const data = await response.json();
        console.log(data);
      } else {
        console.error('Error:', response.status);
      }
    } catch (error) {
      console.error(error);
    }
  };

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    setFile(file);
    console.log(file);
  };

  useEffect(() => {
    async function fetchFiles() {
      try {
        const response = await fetch('/api/files');
        if (response.ok) {
          const data = await response.json();
          console.log(data)
          setFileList(data);
        } else {
          console.error('Error:', response.status);
        }
      } catch (error) {
        console.error(error);
      }
    }
    fetchFiles();
  }, []); // Empty dependency array, so it runs only once on mount

  return (
    <div className="container">
      <h2 className="title">Upload File</h2>
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label className="label">File:</label>
          <input className="input" type="file" accept="*" onChange={handleFileChange} />
        </div>
        <button className="button" type="submit">Submit</button>
      </form>
      <h2 className="title">File List</h2>
      <ul>
        {fileList.map((fileItem) => (
          <li key={fileItem.FID}>
            <a href={`/api/download/${fileItem.fid}`}>{fileItem.fileName}</a>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default FileForm;
