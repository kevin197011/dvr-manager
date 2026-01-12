import { useRef, useEffect } from 'react';

function VideoPlayer({ src, autoPlay = false }) {
  const videoRef = useRef(null);

  useEffect(() => {
    if (videoRef.current) {
      // 停止当前播放
      videoRef.current.pause();
      videoRef.current.currentTime = 0;
      
      if (src) {
        const video = videoRef.current;
        
        // 如果设置了自动播放，监听 canplay 事件后播放
        if (autoPlay) {
          const handleCanPlay = () => {
            const playPromise = video.play();
            if (playPromise !== undefined) {
              playPromise.catch(error => {
                // 自动播放可能被浏览器阻止，静默处理
                console.log('自动播放被阻止:', error);
              });
            }
            video.removeEventListener('canplay', handleCanPlay);
          };
          
          video.addEventListener('canplay', handleCanPlay);
          
          // 加载新视频
          video.load();
          
          // 如果视频已经可以播放，直接播放
          if (video.readyState >= 3) {
            handleCanPlay();
          }
        } else {
          // 不自动播放时，只加载视频
          video.load();
        }
      }
    }
  }, [src, autoPlay]);

  // 组件卸载时停止播放
  useEffect(() => {
    return () => {
      if (videoRef.current) {
        videoRef.current.pause();
        videoRef.current.currentTime = 0;
      }
    };
  }, []);

  if (!src) {
    return null;
  }

  return (
    <video
      ref={videoRef}
      controls
      muted
      style={{ width: '100%', maxHeight: '600px' }}
      preload="metadata"
    >
      <source src={src} type="video/mp4" />
      您的浏览器不支持视频播放
    </video>
  );
}

export default VideoPlayer;
